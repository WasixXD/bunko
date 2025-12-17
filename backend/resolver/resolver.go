package resolver

import (
	"bunko/backend/providers"
	"database/sql"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
)

type Resolver struct {
	// Timer to check for new mangas
	CheckMangaTimer time.Ticker
	Database        *sql.DB
	WorkerChannel   chan int
}

func NewResolver(check time.Duration, db *sql.DB) *Resolver {

	return &Resolver{
		CheckMangaTimer: *time.NewTicker(check),
		Database:        db,
		WorkerChannel:   make(chan int),
	}

}

func (r *Resolver) checkNewManga() int {

	var manga_id int
	r.Database.QueryRow("SELECT manga_id FROM mangas WHERE status = 'pending'").Scan(&manga_id)

	if manga_id == 0 {
		return -1
	}

	log.Info("[Resolver] Found a new manga to download", "manga_id", manga_id)
	r.Database.Exec("UPDATE mangas SET status = 'downloading' WHERE manga_id = ?", manga_id)
	return manga_id
}

// TODO: Refactor make the core functionalities more focused
func (r *Resolver) findChapters(manga_id int) {

	var providerName, url string
	sql := `
        SELECT provider, url
        FROM mangas
        WHERE manga_id = ?
    `
	err := r.Database.QueryRow(sql, manga_id).Scan(&providerName, &url)

	if err != nil {
		log.Error("[Resolver.findChapters()] got error", "error", err)
		return
	}

	factory := providers.NewProviderFactory()
	provider := factory.Get(providerName)

	chapters, err := provider.GetAllChapters(url)

	if err != nil {
		log.Error("[Resolver.findChapters()] got error", "error", err)
		return
	}

	// Maybe this should be part of the db.operations package
	insertIntoQueue := `
		INSERT INTO download_queue(manga_id, name, url, status, provider)
		VALUES (?, ?, ?, 'pending', ?)
	`

	for _, chapter := range chapters {
		_, err := r.Database.Exec(insertIntoQueue, manga_id, chapter.Name, chapter.Url, providerName)

		if err != nil {
			log.Error("[Resolver.insertIntoQueue] got error", "error", err)
			return
		}
	}

	log.Info(fmt.Sprintf("[Resolver] Done! Added %d chapters into download queue", len(chapters)+1))
}

func (r *Resolver) Work() {

	for {
		select {
		case <-r.CheckMangaTimer.C:
			manga_id := r.checkNewManga()
			if manga_id == -1 {
				continue
			}

			// Find Manga Chapters
			r.findChapters(manga_id)

			r.WorkerChannel <- manga_id
		}

	}

}
