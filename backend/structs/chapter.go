package structs

type Chapter struct {
	// Name of the chapter
	Name string

	// Url to find the chapter
	Url string

	Provider string
}

type ChapterJobs struct {
	RowId          int    `json:"rowid"`
	MangaId        int    `json:"manga_id"`
	Name           string `json:"name"`
	Url            string `json:"url"`
	Status         string `json:"status"`
	Provider       string `json:"provider"`
	PathToDownload string `json:"path_to_download"`
}
