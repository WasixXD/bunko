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
)

type Resolver struct {
	// Timer to check for new mangas
	CheckMangaTimer time.Ticker

	UpdateStatusTimer time.Ticker
	Database          *sql.DB
	Downloaders       *downloader.DownloaderLock
}

func NewResolver(check time.Duration, db *sql.DB) *Resolver {

	return &Resolver{
		CheckMangaTimer:   *time.NewTicker(check),
		UpdateStatusTimer: *time.NewTicker(time.Duration(2000 * time.Millisecond)),
		Database:          db,
		Downloaders:       downloader.NewDownloaderBy(50, db),
	}

}

func (r *Resolver) checkNewManga() *structs.Manga {

	var manga structs.Manga
	// TODO: Could be a db.operation
	r.Database.QueryRow("SELECT manga_id, name, manga_path, provider, status, url FROM mangas WHERE status = 'pending'").
		Scan(&manga.MangaId, &manga.Name, &manga.MangaPath, &manga.Provider, &manga.Status, &manga.Url)

	if manga.MangaId == 0 {
		return nil
	}

	log.Info("[Resolver] Found a new manga to download", "manga_id", manga.MangaId)
	return &manga
}

// TODO: Refactor make the structs functionalities more focused
func (r *Resolver) findChapters(manga_id int) ([]structs.Chapter, error) {

	var providerName, url string
	sql := `
        SELECT provider, url
        FROM mangas
        WHERE manga_id = ?
    `

	if err := r.Database.QueryRow(sql, manga_id).Scan(&providerName, &url); err != nil {
		log.Error("[Resolver.findChapters()] got error", "error", err)
		return nil, err
	}

	factory := providers.NewProviderFactory()
	provider := factory.Get(providerName)

	return provider.GetAllChapters(url)

}

// Seems likely that anilist just set the covers as jpg
func (r *Resolver) downloadCover(manga_id int, manga_path, url string) error {

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

func (r *Resolver) notifyWorkers() {
	r.Downloaders.Mutex.Lock()
	r.Downloaders.Cond.Broadcast()
	r.Downloaders.Mutex.Unlock()
}

func (r *Resolver) persistMangaData(
	manga *structs.Manga,
	metadata *structs.AnilistMetadataResponse,
	chapters []structs.Chapter,
) error {

	tx, err := r.Database.Begin()
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

func (r *Resolver) prepareFilesystem(manga *structs.Manga) error {
	dir := filepath.Dir(*manga.MangaPath + "/")
	return os.MkdirAll(dir, 0755)
}

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

func (r *Resolver) UpdateStatus() error {
	mangas, err := db.GetAllMangas(r.Database)
	if err != nil {
		return err
	}

	// TODO: turn into a db.operation
	const query = `
		SELECT * 
		FROM download_queue 
		WHERE manga_id = ?
		AND status = 'pending'
	`
	for _, manga := range mangas {
		err := r.Database.QueryRow(query, manga.MangaId).Scan()
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

func (r *Resolver) Work() {
	for {
		select {
		case <-r.CheckMangaTimer.C:
			if err := r.processNextManga(); err != nil {
				log.Warn(err)
			}
		case <-r.UpdateStatusTimer.C:
			if err := r.UpdateStatus(); err != nil && err != sql.ErrNoRows {
				log.Warn(err)
			}

		}
	}
}
