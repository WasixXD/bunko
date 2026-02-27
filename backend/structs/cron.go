package structs

import "time"

type Cron struct {
	MangaId    int
	Rule       string
	LastUpdate *time.Time
}
