package downloader

import (
	"archive/zip"
	"bunko/backend/core"
	"bunko/backend/providers"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
		WHERE rowid = (
			SELECT rowid
			FROM download_queue
			WHERE status = 'pending'
			ORDER BY manga_id DESC
			LIMIT 1
		)
		RETURNING rowid, name, url, provider, path_to_download;
    `
	row := tx.QueryRow(sqlUpdate)

	err = row.Scan(
		&jb.RowId,
		&jb.Name,
		&jb.Url,
		&jb.Provider,
		&jb.PathToDownload,
	)

	if err != nil {
		return nil, err
	}

	return jb, tx.Commit()
}

func (d *Downloader) SetAsCompleted(download_id int) error {
	tx, err := d.Database.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	sql := `
		UPDATE download_queue
		SET status = 'completed'
		WHERE rowid = ?
	`
	_, err = tx.Exec(sql, download_id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Downloader) TurnIntoCbz(source string) error {
	target := fmt.Sprintf("%s.cbz", source)

	zip_file, err := os.Create(target)
	if err != nil {
		return err
	}

	defer zip_file.Close()

	archive := zip.NewWriter(zip_file)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)

		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (d *Downloader) Run(worker_id int) {

	for {
		chapter, err := d.ClaimChapter()

		if err == sql.ErrNoRows || chapter == nil {
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

		log.Info(fmt.Sprintf("[Downloader:%d] downloading ", worker_id), "chapter", chapter.Name)

		dir := filepath.Dir(chapter.PathToDownload + "/")

		if err = os.MkdirAll(dir, 0755); err != nil {
			log.Warn(err)
			return
		}

		factory := providers.NewProviderFactory()
		provider := factory.Get(chapter.Provider)

		provider.DownloadChapter(chapter.Url, chapter.PathToDownload, chapter.Name)

		old_path := fmt.Sprintf("%s/../cover.jpg", chapter.PathToDownload)
		new_path := fmt.Sprintf("%s/0.jpg", chapter.PathToDownload)
		if err = copyFile(old_path, new_path); err != nil {
			log.Warn(err)
			return
		}

		if err = d.TurnIntoCbz(chapter.PathToDownload); err != nil {
			log.Warn(err)
			return
		}

		if err = d.SetAsCompleted(chapter.RowId); err != nil {
			log.Warn(err)
			return
		}

		log.Info("[Downloader] Done!")
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
