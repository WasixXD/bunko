package resolver

import (
	"bunko/backend/db"

	"github.com/charmbracelet/log"
)

func (r *Resolver) recoverInterruptedWork() error {
	resetJobs, err := db.ResetInterruptedJobs(r.Database)
	if err != nil {
		return err
	}

	if err := r.updateStatus(); err != nil {
		return err
	}

	pendingJobs, err := db.CountPendingJobs(r.Database)
	if err != nil {
		return err
	}

	if resetJobs > 0 {
		log.Info("[Resolver.recoverInterruptedWork] reset interrupted jobs", "count", resetJobs)
	}

	if pendingJobs > 0 {
		log.Info("[Resolver.recoverInterruptedWork] redistributing pending jobs", "count", pendingJobs)
		r.notifyWorkers()
	}

	return nil
}
