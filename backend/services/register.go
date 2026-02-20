package services

import (
	"database/sql"
)

type Services struct {
	Manga *MangaService
	Queue *QueueService
}

func NewServices(database *sql.DB) *Services {
	return &Services{
		Manga: &MangaService{db: database},
		Queue: &QueueService{db: database},
	}
}
