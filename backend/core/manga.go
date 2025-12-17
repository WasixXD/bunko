package core

import "strings"

type Manga struct {
	// Manga name
	Name string

	// Manga name in a snake-case way
	Slug string

	// If the manga is completed
	Status string

	// Indiviual Chapters of that manga
	Chapters []*Chapter

	// The amount of chapters
	ChaptersAmount int
}

func NormalizeName(name string) string {
	newName := []string{}
	for i := range name {
		char := string(name[i])
		if char == " " {
			char = "_"
		}
		char = strings.ToLower(char)
		newName = append(newName, char)

	}
	return strings.Join(newName, "")
}

// Struct to refer to tiny metadata that the backend will hold while the user is
// choosing the manga, none of this columns may exist
type MangaProps struct {
	Name      string `json:"name"`
	Year      string `json:"year"`
	CoverPath string `json:"cover_path"`
	Status    string `json:"status"`
	Url       string `json:"url"`
}
