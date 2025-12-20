package core

import "strings"

// This struct represents the main information from database mangas table
type Manga struct {
	MangaId int

	Name string

	Status string

	Provider string

	Url string

	Path string
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
