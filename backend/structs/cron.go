package structs

import "time"

type Cron struct {
	MangaId    int        `db:"manga_id"`
	Rule       string     `db:"rule"`
	LastUpdate *time.Time `db:"last_updated_at"`
}
