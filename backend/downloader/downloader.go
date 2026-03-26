package downloader

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
)

type Downloader struct {
	Database *sqlx.DB
	Mutex    *sync.Mutex
	Cond     *sync.Cond
}

type DownloaderLock struct {
	Mutex    *sync.Mutex
	Cond     *sync.Cond
	NWorkers int
}

const maxJobRetries = 3

func NewDownloader(database *sqlx.DB, mutex *sync.Mutex, cond *sync.Cond) *Downloader {
	return &Downloader{
		Database: database,
		Mutex:    mutex,
		Cond:     cond,
	}
}

func NewDownloaderBy(n int, database *sqlx.DB) *DownloaderLock {
	newMutex := sync.Mutex{}
	newCond := sync.NewCond(&newMutex)

	for i := range n {
		worker := NewDownloader(database, &newMutex, newCond)
		go worker.Run(i)
	}

	return &DownloaderLock{
		NWorkers: n,
		Cond:     newCond,
		Mutex:    &newMutex,
	}
}

func (d *Downloader) waitForWork() {
	d.Mutex.Lock()
	d.Cond.Wait()
	d.Mutex.Unlock()
}

func (d *Downloader) notifyWorkers() {
	d.Mutex.Lock()
	d.Cond.Broadcast()
	d.Mutex.Unlock()
}

func isNoRows(err error) bool {
	return err == sql.ErrNoRows
}
