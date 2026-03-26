package resolver

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
)

func (r *Resolver) updateStatus() error {
	mangas, err := db.GetAllMangas(r.Database)
	if err != nil {
		return err
	}

	const query = `
		SELECT 1
		FROM download_queue 
		WHERE manga_id = ?
		AND status = 'pending'
		LIMIT 1
	`

	for _, manga := range mangas {
		var pending int
		err := r.Database.Get(&pending, query, manga.MangaId)
		if err == sql.ErrNoRows && manga.Status != "completed" {
			log.Info(fmt.Sprintf("[Resolver] manga %s set as completed", manga.Name))
			db.SetMangaCompleted(r.Database, manga.MangaId)
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) diffChapter(mangaID int) error {
	manga, err := db.GetMangaById(r.Database, fmt.Sprintf("%d", mangaID))
	if err != nil {
		return err
	}

	if manga.MangaPath == nil {
		return fmt.Errorf("manga %d does not have manga_path configured", mangaID)
	}

	chapters, err := r.findChapters(mangaID)
	if err != nil {
		return err
	}

	queuedNames, err := db.GetQueuedChapterNames(r.Database, mangaID)
	if err != nil {
		return err
	}

	existing := make(map[string]struct{}, len(queuedNames))
	for _, name := range queuedNames {
		existing[name] = struct{}{}
	}

	missing := make([]structs.Chapter, 0)
	for _, chapter := range chapters {
		if _, ok := existing[chapter.Name]; ok {
			continue
		}
		missing = append(missing, chapter)
	}

	if len(missing) == 0 {
		log.Info("[Resolver.diffChapter] no new chapters found", "manga_id", mangaID)
		return nil
	}

	tx, err := r.Database.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = db.AddChaptersToQueue(tx, mangaID, *manga.MangaPath, missing); err != nil {
		return err
	}

	if _, err = tx.Exec("UPDATE mangas SET status = 'downloading' WHERE manga_id = ?", mangaID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	log.Info(
		"[Resolver.diffChapter] enqueued new chapters",
		"manga_id", mangaID,
		"count", len(missing),
	)
	r.notifyWorkers()
	return nil
}

func (r *Resolver) checkForTimeRules() error {
	crons, err := db.GetAllTimeRules(r.Database)
	if err != nil {
		return err
	}

	current := make(map[string]string, len(crons))
	for _, cron := range crons {
		tag := fmt.Sprintf("%d", cron.MangaId)
		current[tag] = cron.Rule
	}

	for tag := range r.TimeRules {
		if _, ok := current[tag]; ok {
			continue
		}
		r.Scheduler.RemoveByTags(tag)
		delete(r.TimeRules, tag)
	}

	for _, cron := range crons {
		tag := fmt.Sprintf("%d", cron.MangaId)
		if rule, ok := r.TimeRules[tag]; ok && rule == cron.Rule {
			continue
		}

		if _, ok := r.TimeRules[tag]; ok {
			r.Scheduler.RemoveByTags(tag)
		}

		if _, err := r.Scheduler.NewJob(
			gocron.CronJob(cron.Rule, false),
			gocron.NewTask(r.diffChapter, cron.MangaId),
			gocron.WithTags(tag),
		); err != nil {
			return err
		}

		r.TimeRules[tag] = cron.Rule
	}

	return nil
}
