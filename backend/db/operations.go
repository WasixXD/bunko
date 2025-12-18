package db

import (
	"bunko/backend/core"
	"bunko/backend/server"
	"database/sql"
	"fmt"
)

func AddMangaToDB(db *sql.DB, manga server.MangaPost) (int, error) {

	var manga_id int
	// ./manga_path/manga_name
	absPath := fmt.Sprintf("%s/%s", manga.MangaPath, manga.Name)
	slug := core.NormalizeName(manga.Name)
	// TODO: Make this a transaction?
	sql := `INSERT INTO 
				mangas(name, slug, provider, status, url, manga_path) 
			VALUES (?, ?, ?, 'pending', ?, ?) 
			RETURNING manga_id`
	err := db.QueryRow(sql, manga.Name, slug, manga.ProviderName, manga.Url, absPath).Scan(&manga_id)

	return manga_id, err
}
