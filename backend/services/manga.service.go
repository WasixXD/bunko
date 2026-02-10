package services

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"database/sql"
)

type MangaService struct {
	db *sql.DB
}

func (m *MangaService) AddManga(manga structs.MangaPost) (int, error) {
	return db.AddMangaToDB(m.db, manga)
}
