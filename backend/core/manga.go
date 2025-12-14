package core

type Manga struct {
	// Manga name
	Name string

	// Manga name in a snake-case way
	Slug string

	// If the manga is completed
	Status string

	// The provider of this manga
	Provider *Provider

	// Indiviual Chapters of that manga
	Chapters []*Chapter

	// The amount of chapters
	ChaptersAmount int
}
