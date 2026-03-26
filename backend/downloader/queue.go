package downloader

import (
	"bunko/backend/structs"
	"context"
	"database/sql"

	"github.com/charmbracelet/log"
)

func (d *Downloader) ClaimChapter() (*structs.ChapterJobs, error) {
	tx, err := d.Database.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	job := &structs.ChapterJobs{}
	const query = `
		UPDATE download_queue
		SET status = 'downloading'
		WHERE rowid = (
			SELECT rowid
			FROM download_queue
			WHERE status = 'pending'
			ORDER BY manga_id DESC
			LIMIT 1
		)
		RETURNING rowid, manga_id, name, url, status, provider, path_to_download, retry_count, COALESCE(last_error, '') AS last_error;
	`

	if err := tx.Get(job, query); err != nil {
		return nil, err
	}

	return job, tx.Commit()
}

func (d *Downloader) SetAsCompleted(downloadID int) error {
	tx, err := d.Database.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	const query = `
		UPDATE download_queue
		SET status = 'completed',
			last_error = NULL
		WHERE rowid = ?
	`

	if _, err := tx.Exec(query, downloadID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Downloader) HandleFailedJob(downloadID int, failure error) error {
	tx, err := d.Database.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var job structs.ChapterJobs
	if err := tx.Get(&job, `SELECT rowid, retry_count FROM download_queue WHERE rowid = ?`, downloadID); err != nil {
		return err
	}

	nextRetryCount := job.RetryCount + 1
	nextStatus := "pending"
	if nextRetryCount > maxJobRetries {
		nextStatus = "error"
	}

	if _, err := tx.Exec(
		`UPDATE download_queue
		 SET status = ?, retry_count = ?, last_error = ?
		 WHERE rowid = ?`,
		nextStatus,
		nextRetryCount,
		failure.Error(),
		downloadID,
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if nextStatus == "error" {
		log.Error("[Downloader] job exhausted retries", "rowid", downloadID, "retries", nextRetryCount, "error", failure)
		return nil
	}

	log.Warn("[Downloader] job failed, retrying", "rowid", downloadID, "retry", nextRetryCount, "error", failure)
	d.notifyWorkers()
	return nil
}
