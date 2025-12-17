package db

import (
	"bunko/backend/core"
	"bunko/backend/server"
	"database/sql"
)

func AddMangaToDB(db *sql.DB, manga server.MangaPost) (int, error) {

	var manga_id int
	slug := core.NormalizeName(manga.Name)
	// TODO: Make this a transaction?
	sql := `INSERT INTO 
				mangas(name, slug, provider, status, url) 
			VALUES (?, ?, ?, 'pending', ?) 
			RETURNING manga_id`
	err := db.QueryRow(sql, manga.Name, slug, manga.ProviderName, manga.Url).Scan(&manga_id)

	return manga_id, err
}
