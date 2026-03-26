package services

import (
	"bunko/backend/db"
	"bunko/backend/downloader"
	"bunko/backend/structs"

	"github.com/jmoiron/sqlx"
)

type QueueService struct {
	db          *sqlx.DB
	downloaders *downloader.DownloaderLock
}

func (q *QueueService) GetAll() ([]structs.ChapterJobs, error) {
	return db.GetAllJobs(q.db)
}

func (q *QueueService) Retry(id string) error {
	var rowID int
	if err := q.db.Get(&rowID, `SELECT rowid FROM download_queue WHERE rowid = ?`, id); err != nil {
		return err
	}

	if _, err := q.db.Exec(
		`UPDATE download_queue
		 SET status = 'pending', retry_count = 0, last_error = NULL
		 WHERE rowid = ?`,
		rowID,
	); err != nil {
		return err
	}

	q.downloaders.Mutex.Lock()
	q.downloaders.Cond.Broadcast()
	q.downloaders.Mutex.Unlock()
	return nil
}
