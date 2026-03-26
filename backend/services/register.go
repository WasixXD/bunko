package services

import (
	"github.com/jmoiron/sqlx"
)

type Services struct {
	Manga *MangaService
	Queue *QueueService
}

func NewServices(database *sqlx.DB) *Services {
	return &Services{
		Manga: &MangaService{db: database},
		Queue: &QueueService{db: database},
	}
}
