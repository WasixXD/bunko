package resolver

import (
	"bunko/backend/core"
	"bunko/backend/db"
	"bunko/backend/downloader"
	"bunko/backend/providers"
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
	Database        *sql.DB
	Downloaders     *downloader.DownloaderLock
}

func NewResolver(check time.Duration, db *sql.DB) *Resolver {

	return &Resolver{
		CheckMangaTimer: *time.NewTicker(check),
		Database:        db,
		Downloaders:     downloader.NewDownloaderBy(1, db),
	}

}

func (r *Resolver) checkNewManga() *core.Manga {

	var manga core.Manga
	r.Database.QueryRow("SELECT manga_id, name, manga_path, provider, status, url FROM mangas WHERE status = 'pending'").
		Scan(&manga.MangaId, &manga.Name, &manga.Path, &manga.Provider, &manga.Status, &manga.Url)

	if manga.MangaId == 0 {
		return nil
	}

	log.Info("[Resolver] Found a new manga to download", "manga_id", manga.MangaId)
	return &manga
}

// TODO: Refactor make the core functionalities more focused
func (r *Resolver) findChapters(manga_id int) ([]core.Chapter, error) {

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
func (r *Resolver) downloadCover(manga_path, url string) error {

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

	return nil
}

func (r *Resolver) Work() {

	for {
		select {
		// 1. Check if a new Manga was Added to Database
		// 2. Get Metadata of the manga and save on the Database
		// 3. Create Directory
		// 4. Get Cover Image
		// 5. Add Jobs to Database
		// 6. Notify Workers
		case <-r.CheckMangaTimer.C:
			//
			// TODO: Use transaction in case that the errors happened after operations
			manga := r.checkNewManga()

			if manga == nil {
				continue
			}

			dir := filepath.Dir(manga.Path + "/")

			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Warn(err)
				continue
			}

			metadata, err := core.AnilistMetadataQuery(manga.Name)

			if err != nil {
				log.Warn(err)
				continue
			}

			if err = r.downloadCover(manga.Path, metadata.Data.Media.CoverImage.ExtraLarge); err != nil {
				log.Warn(err)
				continue
			}

			if err = db.AddMetadataToManga(r.Database, manga.MangaId, manga.Url, *metadata); err != nil {
				log.Warn(err)
				continue
			}

			chapters, err := r.findChapters(manga.MangaId)

			if err != nil {
				log.Warn(err)
				continue
			}

			if err = db.AddChaptersToQueue(r.Database, manga.MangaId, manga.Path, chapters); err != nil {
				log.Warn(err)
				continue
			}

			r.Database.Exec("UPDATE mangas SET status = 'downloading' WHERE manga_id = ?", manga.MangaId)

			// Notify workers
			r.Downloaders.Mutex.Lock()
			r.Downloaders.Cond.Broadcast()
			r.Downloaders.Mutex.Unlock()
		}
	}

}
