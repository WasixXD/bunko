package services

import (
	"bunko/backend/db"
	"bunko/backend/structs"

	"github.com/jmoiron/sqlx"
)

type MangaService struct {
	db *sqlx.DB
}

func (m *MangaService) AddManga(manga structs.MangaPost) (int, error) {
	return db.AddMangaToDB(m.db, manga)
}

func (m *MangaService) GetAll() ([]structs.Manga, error) {
	return db.GetAllMangas(m.db)
}

func (m *MangaService) GetById(id string) (structs.Manga, error) {
	return db.GetMangaById(m.db, id)
}

func (m *MangaService) DeleteById(id string) (int, error) {
	return db.DeleteMangaById(m.db, id)
}

func (m *MangaService) AddTimeRule(time_rule, manga_id string) error {
	return db.AddTimeRule(m.db, time_rule, manga_id)
}
