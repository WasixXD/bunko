package resolver

import (
	"bunko/backend/downloader"
	"bunko/backend/providers"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jmoiron/sqlx"
)

type Resolver struct {
	// Timer to check for new mangas
	CheckMangaTimer time.Ticker

	// Timer to update the status of a manga
	UpdateStatusTimer time.Ticker

	// Timer to check for new time rules
	CheckTimeRulesTimer time.Ticker

	// Timer to cleanup folders for mangas deleted from the database
	CleanupDeletedFoldersTimer time.Ticker

	Database    *sqlx.DB
	Downloaders *downloader.DownloaderLock

	Scheduler gocron.Scheduler
	TimeRules map[string]string
	Provider  func(string) providers.Provider
}

func NewResolver(check time.Duration, db *sqlx.DB) *Resolver {
	s, _ := gocron.NewScheduler()

	return &Resolver{
		CheckMangaTimer:            *time.NewTicker(check),
		UpdateStatusTimer:          *time.NewTicker(2000 * time.Millisecond),
		CheckTimeRulesTimer:        *time.NewTicker(10_000 * time.Millisecond),
		CleanupDeletedFoldersTimer: *time.NewTicker(10_000 * time.Millisecond),
		Database:                   db,
		Downloaders:                downloader.NewDownloaderBy(50, db),
		Scheduler:                  s,
		TimeRules:                  map[string]string{},
		Provider: func(name string) providers.Provider {
			return providers.NewProviderFactory().Get(name)
		},
	}
}
