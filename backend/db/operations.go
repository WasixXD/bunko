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

func AddChaptersToQueue(
	tx *sql.Tx,
	mangaID int,
	mangaPath string,
	chapters []core.Chapter,
) error {

	const query = `
		INSERT INTO download_queue
			(manga_id, name, url, status, provider, path_to_download)
		VALUES
			(?, ?, ?, 'pending', ?, ?)
	`

	for _, chapter := range chapters {
		pathToDownload := fmt.Sprintf("%s/%s", mangaPath, chapter.Name)

		if _, err := tx.Exec(
			query,
			mangaID,
			chapter.Name,
			chapter.Url,
			chapter.Provider,
			pathToDownload,
		); err != nil {
			log.Error(
				"[db.AddChaptersToQueue] failed to insert chapter",
				"chapter", chapter.Name,
				"error", err,
			)
			return err
		}
	}

	log.Info(
		"[db.AddChaptersToQueue] chapters enqueued",
		"count", len(chapters),
		"manga_id", mangaID,
	)

	return nil
}

func AddMetadataToManga(
	tx *sql.Tx,
	mangaID int,
	mangaURL string,
	metadata core.AnilistMetadataResponse,
) error {

	media := metadata.Data.Media

	const query = `
		UPDATE mangas
		SET
			localized_name = ?,
			publication_status = ?,
			summary = ?,
			start_year = ?,
			start_month = ?,
			start_day = ?,
			web_link = ?,
			metadata_updated_at = datetime('now')
		WHERE manga_id = ?
	`

	_, err := tx.Exec(
		query,
		media.Title.Native,
		media.Status,
		media.Description,
		media.StartDate.Year,
		media.StartDate.Month,
		media.StartDate.Day,
		mangaURL,
		mangaID,
	)

	if err != nil {
		log.Error(
			"[db.AddMetadataToManga] failed to update metadata",
			"manga_id", mangaID,
			"error", err,
		)
	}

	return err
}
