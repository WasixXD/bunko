package resolver

import (
	"database/sql"

	"github.com/charmbracelet/log"
)

func (r *Resolver) Work() {
	r.Scheduler.Start()
	for {
		select {
		case <-r.CheckMangaTimer.C:
			if err := r.processNextManga(); err != nil {
				log.Warn(err)
			}
		case <-r.UpdateStatusTimer.C:
			if err := r.updateStatus(); err != nil && err != sql.ErrNoRows {
				log.Warn(err)
			}
		case <-r.CheckTimeRulesTimer.C:
			if err := r.checkForTimeRules(); err != nil {
				log.Warn(err)
			}
		}
	}
}
