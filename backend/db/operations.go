package db

import (
	"bunko/backend/structs"
	"database/sql"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
)

const mangaSelectColumns = `
	SELECT
		m.manga_id,
		m.name,
		m.slug,
		m.status,
		m.provider,
		m.url,
		m.cover_path,
		m.manga_path,
		md.localized_name,
		md.publication_status,
		md.summary,
		md.start_year,
		md.start_month,
		md.start_day,
		md.author,
		md.art,
		md.web_link,
		md.metadata_updated_at,
		m.created_at
	FROM mangas m
	LEFT JOIN manga_metadata md ON md.manga_id = m.manga_id
`

func AddMangaToDB(db *sqlx.DB, manga structs.MangaPost) (int, error) {

	var manga_id int
	// ./manga_path/manga_name
	absPath := filepath.Join(manga.MangaPath, manga.Name)
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
	var mangas []structs.Manga
	if err := db.Select(&mangas, mangaSelectColumns); err != nil {
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
		pathToDownload := filepath.Join(mangaPath, chapter.Name)

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
	author, art := metadata.Creators()

	const query = `
		INSERT INTO manga_metadata (
			manga_id,
			localized_name,
			publication_status,
			summary,
			start_year,
			start_month,
			start_day,
			author,
			art,
			web_link,
			metadata_updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(manga_id) DO UPDATE SET
			localized_name = excluded.localized_name,
			publication_status = excluded.publication_status,
			summary = excluded.summary,
			start_year = excluded.start_year,
			start_month = excluded.start_month,
			start_day = excluded.start_day,
			author = excluded.author,
			art = excluded.art,
			web_link = excluded.web_link,
			metadata_updated_at = datetime('now')
	`

	_, err := tx.Exec(
		query,
		mangaID,
		media.Title.Native,
		media.Status,
		media.Description,
		media.StartDate.Year,
		media.StartDate.Month,
		media.StartDate.Day,
		nullIfEmpty(author),
		nullIfEmpty(art),
		mangaURL,
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
	return SetMangaStatus(db, manga_id, "completed")
}

func SetMangaStatus(db *sqlx.DB, mangaID int, status string) error {
	const query = `
		UPDATE mangas 
		SET status = ?
		WHERE manga_id = ?
	`
	_, err := db.Exec(query, status, mangaID)
	if err != nil {
		return err
	}

	return nil
}

func SetMangaCoverPath(tx *sqlx.Tx, mangaID int, coverPath string) error {
	const query = `
		UPDATE mangas
		SET cover_path = ?
		WHERE manga_id = ?
	`

	_, err := tx.Exec(query, coverPath, mangaID)
	return err
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
	query := mangaSelectColumns + `
		WHERE m.manga_id = ?
	`
	var manga structs.Manga

	err := db.Get(&manga, query, id)

	if err != nil {
		return manga, err
	}

	return manga, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func DeleteMangaById(db *sqlx.DB, id string) (int, error) {
	const selectQuery = `
		SELECT manga_path
		FROM mangas
		WHERE manga_id = ?
	`
	const deleteQuery = `
		DELETE FROM mangas
		WHERE manga_id = ? 
	`

	tx, err := db.Beginx()
	if err != nil {
		return -1, err
	}

	defer tx.Rollback()

	var mangaPath string
	if err := tx.Get(&mangaPath, selectQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return -1, err
	}

	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO deleted_manga_paths(path) VALUES (?)`,
		mangaPath,
	); err != nil {
		return -1, err
	}

	result, err := tx.Exec(deleteQuery, id)

	if err != nil {
		return -1, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}

	return int(rows), nil
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

func GetQueuedChapterNames(db *sqlx.DB, mangaID int) ([]string, error) {
	const query = `
		SELECT name
		FROM download_queue
		WHERE manga_id = ?
	`

	var names []string
	if err := db.Select(&names, query, mangaID); err != nil {
		return nil, err
	}

	return names, nil
}

func GetAllTimeRules(db *sqlx.DB) ([]structs.Cron, error) {
	const query = `
		SELECT manga_id, rule, last_updated_at
		FROM cron
	`

	var crons []structs.Cron
	if err := db.Select(&crons, query); err != nil {
		return nil, err
	}

	return crons, nil
}

func GetMangaQueueStatusCounts(db *sqlx.DB, mangaID int) (map[string]int, error) {
	const query = `
		SELECT status, COUNT(*) AS total
		FROM download_queue
		WHERE manga_id = ?
		GROUP BY status
	`

	var rows []struct {
		Status string `db:"status"`
		Total  int    `db:"total"`
	}

	if err := db.Select(&rows, query, mangaID); err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.Total
	}

	return counts, nil
}

func GetDeletedMangaPaths(db *sqlx.DB) ([]string, error) {
	const query = `
		SELECT path
		FROM deleted_manga_paths
		ORDER BY created_at ASC
	`

	var paths []string
	if err := db.Select(&paths, query); err != nil {
		return nil, err
	}

	return paths, nil
}

func DeleteQueuedMangaPath(db *sqlx.DB, path string) error {
	const query = `
		DELETE FROM deleted_manga_paths
		WHERE path = ?
	`

	_, err := db.Exec(query, path)
	return err
}
