package services

import (
	"bunko/backend/db"
	"bunko/backend/downloader"
	"bunko/backend/structs"
	"database/sql"
	"strconv"

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
	exists, err := db.QueueJobExists(q.db, id)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	rowID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	if err := db.ResetQueueJobForRetry(q.db, rowID); err != nil {
		return err
	}

	q.downloaders.Mutex.Lock()
	q.downloaders.Cond.Broadcast()
	q.downloaders.Mutex.Unlock()
	return nil
}
