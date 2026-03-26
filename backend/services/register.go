package services

import (
	"bunko/backend/downloader"

	"github.com/jmoiron/sqlx"
)

type Services struct {
	Manga *MangaService
	Queue *QueueService
}

func NewServices(database *sqlx.DB, downloaders *downloader.DownloaderLock) *Services {
	return &Services{
		Manga: &MangaService{db: database},
		Queue: &QueueService{db: database, downloaders: downloaders},
	}
}
