package resolver

import (
	"bunko/backend/db"
	"bunko/backend/downloader"
	"bunko/backend/providers"
	"bunko/backend/structs"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
)

type Resolver struct {
	// Timer to check for new mangas
	CheckMangaTimer time.Ticker

	// Timer to update the status of a manga
	UpdateStatusTimer time.Ticker

	// Timer to check for new time rules
	CheckTimeRulesTimer time.Ticker

	Database    *sqlx.DB
	Downloaders *downloader.DownloaderLock

	Scheduler gocron.Scheduler
	TimeRules []string
}

func NewResolver(check time.Duration, db *sqlx.DB) *Resolver {
	s, _ := gocron.NewScheduler()

	return &Resolver{
		CheckMangaTimer:     *time.NewTicker(check),
		UpdateStatusTimer:   *time.NewTicker(time.Duration(2000 * time.Millisecond)),
		CheckTimeRulesTimer: *time.NewTicker(time.Duration(10_000 * time.Millisecond)),
		Database:            db,
		Downloaders:         downloader.NewDownloaderBy(50, db),
		Scheduler:           s,
		TimeRules:           []string{},
	}

}

/*
checkNewManga() *structs.Manga

This functions is responsable for checking if any manga was added into
the database.
*/
func (r *Resolver) checkNewManga() *structs.Manga {

	var manga structs.Manga
	err := r.Database.Get(
		&manga,
		"SELECT manga_id, name, manga_path, provider, status, url FROM mangas WHERE status = 'pending'",
	)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		log.Error("[Resolver.checkNewManga] failed to query pending manga", "error", err)
		return nil
	}

	if manga.MangaId == 0 {
		return nil
	}

	log.Info("[Resolver] Found a new manga to download", "manga_id", manga.MangaId)
	return &manga
}

/*
findChapters(manga_id int) ([]structs.Chapter, error)

This function selects the provider and url from the database
and call the right provider to get chapters on its website.
*/
func (r *Resolver) findChapters(manga_id int) ([]structs.Chapter, error) {

	var mangaSource struct {
		Provider string `db:"provider"`
		URL      string `db:"url"`
	}
	sql := `
        SELECT provider, url
        FROM mangas
        WHERE manga_id = ?
    `

	if err := r.Database.Get(&mangaSource, sql, manga_id); err != nil {
		log.Error("[Resolver.findChapters()] got error", "error", err)
		return nil, err
	}

	factory := providers.NewProviderFactory()
	provider := factory.Get(mangaSource.Provider)

	return provider.GetAllChapters(mangaSource.URL)

}

/*
downloadCover(manga_id int, manga_path, url string) error

This functions receives is resposible to download the cover from
anilist so the workers can use it.
*/
func (r *Resolver) downloadCover(manga_id int, manga_path, url string) error {
	// Seems likely that anilist just set the covers as jpg
	absPath := fmt.Sprintf("%s/cover.jpg", manga_path)

	file, err := os.Create(absPath)
	if err != nil {
		return err
	}

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)

	_, err = file.Write(b)

	if err != nil {
		return err
	}

	r.Database.Exec("UPDATE mangas SET cover_path = ? WHERE manga_id = ?", url, manga_id)

	return nil
}

/*
notifyWorkers()

This functions is responsible for waking up workers so they can
start downloading the chapters.
*/
func (r *Resolver) notifyWorkers() {
	r.Downloaders.Mutex.Lock()
	r.Downloaders.Cond.Broadcast()
	r.Downloaders.Mutex.Unlock()
}

/*
persistMangaData(

	manga *structs.Manga,
	metadata *structs.AnilistMetadataResponse,
	chapters []structs.Chapter,

) error

This functions is a series of steps that the resolver take to add all information
needed right back into the database. This consists of:
1. Settings the metadata into the mangas table.
2. Adding jobs into the download queue.
3. Set the right status for the manga.
*/
func (r *Resolver) persistMangaData(
	manga *structs.Manga,
	metadata *structs.AnilistMetadataResponse,
	chapters []structs.Chapter,
) error {

	tx, err := r.Database.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = db.AddMetadataToManga(tx, manga.MangaId, manga.Url, *metadata); err != nil {
		return err
	}

	if err = db.AddChaptersToQueue(tx, manga.MangaId, *manga.MangaPath, chapters); err != nil {
		return err
	}

	if _, err = tx.Exec("UPDATE mangas SET status = 'downloading' WHERE manga_id = ?", manga.MangaId); err != nil {
		return err
	}

	return tx.Commit()
}

/*
prepareFilesystem() error

This function creates the right folders to where the manga will be downloaded
*/
func (r *Resolver) prepareFilesystem(manga *structs.Manga) error {
	dir := filepath.Dir(*manga.MangaPath + "/")
	return os.MkdirAll(dir, 0755)
}

/*
processNextManga() error

This function is a series of steps the resolver take to start the download
of a manga. This consists of:
1. Checking if any new manga was added.
2. Preparing the filesystem.
3. Get the metadata from anilist.
4. Download the manga cover.
5. Find the chapters available to download.
6. Persist the data into the database.
7. Notify all workers.
*/
func (r *Resolver) processNextManga() error {
	manga := r.checkNewManga()
	if manga == nil {
		return nil
	}

	if err := r.prepareFilesystem(manga); err != nil {
		return err
	}

	metadata, err := structs.AnilistMetadataQuery(manga.Name)
	if err != nil {
		return err
	}

	if err := r.downloadCover(manga.MangaId, *manga.MangaPath, metadata.Data.Media.CoverImage.ExtraLarge); err != nil {
		return err
	}

	chapters, err := r.findChapters(manga.MangaId)
	if err != nil {
		return err
	}

	if err := r.persistMangaData(manga, metadata, chapters); err != nil {
		return err
	}

	r.notifyWorkers()
	return nil
}

/*
updateStatus() error

This is a side job that the resolver takes to check if all downloads
were concluded.
*/
func (r *Resolver) updateStatus() error {
	mangas, err := db.GetAllMangas(r.Database)
	if err != nil {
		return err
	}

	// TODO: turn into a db.operation
	const query = `
		SELECT 1
		FROM download_queue 
		WHERE manga_id = ?
		AND status = 'pending'
		LIMIT 1
	`
	for _, manga := range mangas {
		var pending int
		err := r.Database.Get(&pending, query, manga.MangaId)
		if err == sql.ErrNoRows && manga.Status != "completed" {
			log.Info(fmt.Sprintf("[Resolver] manga %s set as completed", manga.Name))
			db.SetMangaCompleted(r.Database, manga.MangaId)
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) diffChapter() error {
	log.Info("[Resolver.diffChapter()] TODO")
	return nil
}

// TODO: Refactor
func (r *Resolver) checkForTimeRules() error {
	query := `
		SELECT count(*) 
		FROM cron
	`

	var newTimeRules int
	err := r.Database.Get(&newTimeRules, query)
	if err != nil {
		return err
	}

	currentTimeRules := len(r.TimeRules)

	// this will happens most of times
	if newTimeRules == currentTimeRules {
		return nil
	}

	log.Info("[Resolver.checkForTimeRules] Updating rules...")
	// add job
	if newTimeRules > currentTimeRules {
		query = `
			SELECT *
			FROM cron
		`
		var crons []structs.Cron
		if err := r.Database.Select(&crons, query); err != nil {
			return err
		}

		for _, cron := range crons {
			strMangaId := fmt.Sprintf("%d", cron.MangaId)
			r.Scheduler.NewJob(
				gocron.CronJob(cron.Rule, false),
				gocron.NewTask(r.diffChapter),
				gocron.WithTags(strMangaId),
			)
			r.TimeRules = append(r.TimeRules, strMangaId)
		}
		return nil
	}

	if newTimeRules < currentTimeRules {
		log.Info("[Resolver.checkForTimeRules()] newTimeRules < currentTimeRules")
		return nil
	}

	return nil
}

func (r *Resolver) Work() {
	r.Scheduler.Start()
	for {
		select {
		case <-r.CheckMangaTimer.C:
			if err := r.processNextManga(); err != nil {
				log.Warn(err)
			}
		case <-r.UpdateStatusTimer.C:
			if err := r.updateStatus(); err != nil && err != sql.ErrNoRows {
				log.Warn(err)
			}

		case <-r.CheckTimeRulesTimer.C:
			if err := r.checkForTimeRules(); err != nil {
				log.Warn(err)
			}

		}
	}
}
