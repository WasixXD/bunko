package downloader

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"context"

	"github.com/charmbracelet/log"
)

func (d *Downloader) ClaimChapter() (*structs.ChapterJobs, error) {
	return db.ClaimNextQueueJob(d.Database)
}

func (d *Downloader) SetAsCompleted(downloadID int) error {
	return db.SetQueueJobCompleted(d.Database, downloadID)
}

func (d *Downloader) HandleFailedJob(downloadID int, failure error) error {
	tx, err := d.Database.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	retryCount, err := db.GetQueueJobRetryCount(tx, downloadID)
	if err != nil {
		return err
	}

	nextRetryCount := retryCount + 1
	nextStatus := "pending"
	if nextRetryCount > maxJobRetries {
		nextStatus = "error"
	}

	if err := db.UpdateQueueJobFailure(tx, downloadID, nextStatus, nextRetryCount, failure.Error()); err != nil {
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
