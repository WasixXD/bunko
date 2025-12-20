package db

import (
	"bunko/backend/core"
	"bunko/backend/server"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
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

	err := db.QueryRow(sql,
		manga.Name,
		slug,
		manga.ProviderName,
		manga.Url,
		absPath,
	).Scan(&manga_id)

	return manga_id, err
}

func AddChaptersToQueue(db *sql.DB, manga_id int, manga_path string, chapters []core.Chapter) error {
	sql := `
		INSERT INTO download_queue(manga_id, name, url, status, provider, path_to_download)
		VALUES (?, ?, ?, 'pending', ?, ?)
	`

	i := 0
	for _, chapter := range chapters {
		// By now we have ./manga_path/manga_name/chapter_name
		path_to_download := fmt.Sprintf("%s/%s/", manga_path, chapter.Name)

		_, err := db.Exec(
			sql,
			manga_id,
			chapter.Name,
			chapter.Url,
			chapter.Provider,
			path_to_download,
		)

		if err != nil {
			log.Error("[Resolver.insertIntoQueue] got error", "error", err)
			return err
		}
		i++
	}

	log.Info(fmt.Sprintf("Done! Added %d chapters into download queue", i))
	return nil

}

func AddMetadataToManga(db *sql.DB, manga_id int, manga_url string, metadata core.AnilistMetadataResponse) error {

	media := metadata.Data.Media
	// TODO: Solve staff, tags and genres problem
	sql := `
		UPDATE mangas
		SET localized_name = ?,
			publication_status = ?,
			summary = ?,
			start_year = ?,
			start_month = ?,
			start_day = ?,
			web_link = ?,
			metadata_updated_at = datetime('now')
		WHERE manga_id = ?
		`

	_, err := db.Exec(
		sql,
		media.Title.Native,
		media.Status,
		media.Description,
		media.StartDate.Year,
		media.StartDate.Month,
		media.StartDate.Day,
		manga_url,
		manga_id,
	)
	return err

}
