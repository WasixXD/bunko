package services

import (
	"database/sql"
)

type Services struct {
	Manga *MangaService
}

func NewServices(database *sql.DB) *Services {
	return &Services{
		Manga: &MangaService{db: database},
	}
}
