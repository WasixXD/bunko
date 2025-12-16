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
