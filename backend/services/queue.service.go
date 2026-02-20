package services

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"database/sql"
)

type QueueService struct {
	db *sql.DB
}

func (q *QueueService) GetAll() ([]structs.ChapterJobs, error) {
	return db.GetAllJobs(q.db)
}
