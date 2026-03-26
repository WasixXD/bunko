package services

import (
	"bunko/backend/db"
	"bunko/backend/structs"

	"github.com/jmoiron/sqlx"
)

type QueueService struct {
	db *sqlx.DB
}

func (q *QueueService) GetAll() ([]structs.ChapterJobs, error) {
	return db.GetAllJobs(q.db)
}
