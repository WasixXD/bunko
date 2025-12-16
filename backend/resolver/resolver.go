package resolver

import (
	"database/sql"
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

func (r *Resolver) findChapters(manga_id int) {

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

			go r.findChapters(manga_id)

			r.WorkerChannel <- manga_id
		}

	}

}
