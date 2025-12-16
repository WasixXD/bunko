package db

import (
	"bunko/backend/core"
	"bunko/backend/server"
	"database/sql"
)

func AddMangaToDB(db *sql.DB, manga server.MangaPost) error {

	// TODO: Make this a transaction?
	slug := core.NormalizeName(manga.Name)
	_, err := db.Exec(
		"INSERT INTO mangas(name, slug, provider, status) VALUES (?, ?, ?, 'pending') RETURNING manga_id",
		manga.Name, slug, manga.ProviderName,
	)

	return err
}
