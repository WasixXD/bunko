package db

import (
	"bunko/backend/structs"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
)

func AddMangaToDB(db *sqlx.DB, manga structs.MangaPost) (int, error) {

	var manga_id int
	// ./manga_path/manga_name
	absPath := fmt.Sprintf("%s/%s", manga.MangaPath, manga.Name)
	slug := structs.NormalizeName(manga.Name)
	// TODO: Make this a transaction?
	sql := `INSERT INTO 
				mangas(name, slug, provider, status, url, manga_path) 
			VALUES (?, ?, ?, 'pending', ?, ?) 
			RETURNING manga_id`

	err := db.Get(&manga_id, sql,
		manga.Name,
		slug,
		manga.ProviderName,
		manga.Url,
		absPath,
	)

	return manga_id, err
}

func GetAllMangas(db *sqlx.DB) ([]structs.Manga, error) {

	sql := `SELECT * FROM mangas`
	var mangas []structs.Manga
	if err := db.Select(&mangas, sql); err != nil {
		return nil, err
	}

	return mangas, nil
}

func AddChaptersToQueue(
	tx *sqlx.Tx,
	mangaID int,
	mangaPath string,
	chapters []structs.Chapter,
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
	tx *sqlx.Tx,
	mangaID int,
	mangaURL string,
	metadata structs.AnilistMetadataResponse,
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

func SetMangaCompleted(db *sqlx.DB, manga_id int) error {
	const query = `
		UPDATE mangas 
		SET status = 'completed'
		WHERE manga_id = ?
	`
	_, err := db.Exec(query, manga_id)
	if err != nil {
		return err
	}

	return nil
}

func GetAllJobs(db *sqlx.DB) ([]structs.ChapterJobs, error) {
	const query = `
		SELECT rowid, * 
		FROM download_queue
	`
	var jobs []structs.ChapterJobs
	if err := db.Select(&jobs, query); err != nil {
		return nil, err
	}

	return jobs, nil
}

func GetMangaById(db *sqlx.DB, id string) (structs.Manga, error) {
	const query = `
		SELECT *
		FROM mangas
		WHERE manga_id = ?
	`
	var manga structs.Manga

	err := db.Get(&manga, query, id)

	if err != nil {
		return manga, err
	}

	return manga, nil
}

func DeleteMangaById(db *sqlx.DB, id string) (int, error) {
	const query = `
		DELETE FROM mangas
		WHERE manga_id = ? 
	`

	result, err := db.Exec(query, id)

	if err != nil {
		return -1, err
	}

	rows, err := result.RowsAffected()

	return int(rows), err
}

func AddTimeRule(db *sqlx.DB, time_rule, manga_id string) error {
	const query = `
		INSERT INTO cron(manga_id, rule)
		VALUES (?, ?)
	`

	_, err := db.Exec(query, manga_id, time_rule)

	if err != nil {
		return err
	}

	return nil
}
