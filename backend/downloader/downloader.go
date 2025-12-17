package downloader

import (
	"bunko/backend/core"
	"bunko/backend/providers"
	"context"
	"database/sql"
	"sync"

	"github.com/charmbracelet/log"
)

type Downloader struct {
	Database *sql.DB
	Mutex    *sync.Mutex
	Cond     *sync.Cond
}

type DownloaderLock struct {
	Mutex    *sync.Mutex
	Cond     *sync.Cond
	NWorkers int
}

func (d *Downloader) ClaimChapter() (*core.ChapterJobs, error) {

	tx, err := d.Database.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	jb := &core.ChapterJobs{}

	sqlUpdate := `
		UPDATE download_queue
		SET status = 'downloading'
		WHERE manga_id = (
			SELECT manga_id
			FROM download_queue
			WHERE status = 'pending'
			ORDER BY manga_id DESC
			LIMIT 1
		);
    `
	_, err = tx.Exec(sqlUpdate)

	if err != nil {
		return nil, err
	}

	sqlSelect := `
		SELECT manga_id, name, url, status, provider
		FROM download_queue
		WHERE status = 'downloading'
		ORDER BY manga_id
		LIMIT 1;
	`
	row := tx.QueryRow(sqlSelect)

	err = row.Scan(
		&jb.MangaId,
		&jb.Name,
		&jb.Url,
		&jb.Status,
		&jb.Provider,
	)

	if err != nil {
		return nil, err
	}

	return jb, tx.Commit()
}

func (d *Downloader) Run() {

	for {
		job, err := d.ClaimChapter()

		if err == sql.ErrNoRows || job == nil {
			d.Mutex.Lock()
			for {
				d.Cond.Wait()
				break
			}
			d.Mutex.Unlock()
			continue
		}

		// TODO: If we fail, the job continue to be claimed?
		if err != nil {
			log.Error("[Downloader] failed on claim job", "error", err)
			continue
		}

		log.Info("[Downloader] downloading ", "chapter", job.Name)
		factory := providers.NewProviderFactory()
		provider := factory.Get(job.Provider)

		provider.DownloadChapter(job.Url, "mangas", job.Name)

		log.Info("[Downloader] Done!")
		break
	}
}

func NewDownloader(database *sql.DB, mutex *sync.Mutex, cond *sync.Cond) *Downloader {
	return &Downloader{
		Database: database,
		Mutex:    mutex,
		Cond:     cond,
	}
}

func NewDownloaderBy(n int, database *sql.DB) *DownloaderLock {
	newMutex := sync.Mutex{}
	newCond := sync.NewCond(&newMutex)

	for range n {
		worker := NewDownloader(database, &newMutex, newCond)
		go worker.Run()
	}

	return &DownloaderLock{
		NWorkers: n,
		Cond:     newCond,
		Mutex:    &newMutex,
	}

}
