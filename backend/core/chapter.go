package core

type Chapter struct {
	// Name of the chapter
	Name string

	// Manga that the chapter belongs to
	Manga *Manga

	// Url to find the chapter
	Url string

	// Chapter number
	Number int
}
