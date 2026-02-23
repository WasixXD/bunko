package db

import (
	"bunko/backend/structs"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
)

func AddMangaToDB(db *sql.DB, manga structs.MangaPost) (int, error) {

	var manga_id int
	// ./manga_path/manga_name
	absPath := fmt.Sprintf("%s/%s", manga.MangaPath, manga.Name)
	slug := structs.NormalizeName(manga.Name)
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

func GetAllMangas(db *sql.DB) ([]structs.Manga, error) {

	sql := `SELECT * FROM mangas`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []structs.Manga

	for rows.Next() {
		var m structs.Manga
		// TODO: Lets migrate to sqlx to avoid this giants scans
		err := rows.Scan(
			&m.MangaId,
			&m.Name,
			&m.Slug,
			&m.Status,
			&m.Provider,
			&m.Url,
			&m.CoverPath,
			&m.MangaPath,
			&m.LocalizedName,
			&m.PublicationStatus,
			&m.Summary,
			&m.StartYear,
			&m.StartMonth,
			&m.StartDay,
			&m.Author,
			&m.WebLink,
			&m.MetadataUpdatedAt,
			&m.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		mangas = append(mangas, m)
	}
	return mangas, nil
}

func AddChaptersToQueue(
	tx *sql.Tx,
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
	tx *sql.Tx,
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

func SetMangaCompleted(db *sql.DB, manga_id int) error {
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

func GetAllJobs(db *sql.DB) ([]structs.ChapterJobs, error) {
	const query = `
		SELECT rowid, * 
		FROM download_queue
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []structs.ChapterJobs

	for rows.Next() {
		var j structs.ChapterJobs

		err := rows.Scan(
			&j.RowId,
			&j.MangaId,
			&j.Name,
			&j.Url,
			&j.Status,
			&j.Provider,
			&j.PathToDownload,
		)

		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func GetById(db *sql.DB, id string) (structs.Manga, error) {
	const query = `
		SELECT *
		FROM mangas
		WHERE manga_id = ?
	`
	var manga structs.Manga

	err := db.QueryRow(query, id).Scan(
		&manga.MangaId,
		&manga.Name,
		&manga.Slug,
		&manga.Status,
		&manga.Provider,
		&manga.Url,
		&manga.CoverPath,
		&manga.MangaPath,
		&manga.LocalizedName,
		&manga.PublicationStatus,
		&manga.Summary,
		&manga.StartYear,
		&manga.StartMonth,
		&manga.StartDay,
		&manga.Author,
		&manga.WebLink,
		&manga.MetadataUpdatedAt,
		&manga.CreatedAt,
	)

	if err != nil {
		return manga, err
	}

	return manga, nil
}
