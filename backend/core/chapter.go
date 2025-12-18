package core

type Chapter struct {
	// Name of the chapter
	Name string

	// Url to find the chapter
	Url string
}

type ChapterJobs struct {
	RowId          int
	MangaId        int
	Name           string
	Url            string
	Status         string
	Provider       string
	PathToDownload string
}
