package core

type Chapter struct {
	// Name of the chapter
	Name string

	// Url to find the chapter
	Url string
}

type ChapterJobs struct {
	MangaId  int
	Name     string
	Url      string
	Status   string
	Provider string
}
