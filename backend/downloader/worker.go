package downloader

import (
	"bunko/backend/providers"
	"bunko/backend/structs"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func (d *Downloader) Run(workerID int) {
	for {
		job, err := d.ClaimChapter()
		if isNoRows(err) || job == nil {
			d.waitForWork()
			continue
		}

		if err != nil {
			log.Error("[Downloader] failed to claim job", "error", err)
			continue
		}

		if err := d.processJob(workerID, job); err != nil {
			if retryErr := d.HandleFailedJob(job.RowId, err); retryErr != nil {
				log.Error("[Downloader] failed to update job after processing error", "rowid", job.RowId, "error", retryErr)
			}
			continue
		}

		if err := d.SetAsCompleted(job.RowId); err != nil {
			log.Warn("[Downloader] failed to mark job as completed", "rowid", job.RowId, "error", err)
			if retryErr := d.HandleFailedJob(job.RowId, err); retryErr != nil {
				log.Error("[Downloader] failed to update job after completion error", "rowid", job.RowId, "error", retryErr)
			}
			continue
		}

		log.Info("[Downloader] done", "worker_id", workerID, "chapter", job.Name)
	}
}

func (d *Downloader) processJob(workerID int, job *structs.ChapterJobs) error {
	log.Info("[Downloader] downloading chapter", "worker_id", workerID, "chapter", job.Name)

	if err := os.MkdirAll(job.PathToDownload, 0755); err != nil {
		return err
	}

	provider := providers.NewProviderFactory().Get(job.Provider)
	if provider == nil {
		return fmt.Errorf("provider %q not found", job.Provider)
	}

	if err := provider.DownloadChapter(job.Url, job.PathToDownload, job.Name); err != nil {
		return err
	}

	oldPath := filepath.Join(filepath.Dir(job.PathToDownload), "cover.jpg")
	newPath := filepath.Join(job.PathToDownload, "0.jpg")
	if err := copyFile(oldPath, newPath); err != nil {
		return err
	}

	if err := d.CreateComicInfo(job.MangaId, job.Name, job.PathToDownload); err != nil {
		return err
	}

	if err := d.TurnIntoCbz(job.PathToDownload); err != nil {
		return err
	}

	return nil
}

var _ = sql.ErrNoRows
