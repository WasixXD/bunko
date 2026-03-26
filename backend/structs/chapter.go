package structs

type Chapter struct {
	// Name of the chapter
	Name string

	// Url to find the chapter
	Url string

	Provider string
}

type ChapterJobs struct {
	RowId          int    `json:"rowid" db:"rowid"`
	MangaId        int    `json:"manga_id" db:"manga_id"`
	Name           string `json:"name" db:"name"`
	Url            string `json:"url" db:"url"`
	Status         string `json:"status" db:"status"`
	Provider       string `json:"provider" db:"provider"`
	PathToDownload string `json:"path_to_download" db:"path_to_download"`
	RetryCount     int    `json:"retry_count" db:"retry_count"`
	LastError      string `json:"last_error" db:"last_error"`
}
