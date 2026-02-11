package structs

import (
	"strings"
	"time"
)

// This struct represents the main information from database mangas table
type Manga struct {
	MangaId int `json:"manga_id"`

	Name string `json:"name"`
	Slug string `json:"slug"`

	// System status
	Status string `json:"status"`

	Provider string `json:"provider"`
	Url      string `json:"url"`

	CoverPath *string `json:"cover_path"`
	MangaPath *string `json:"manga_path"`

	LocalizedName     *string    `json:"localized_name"`
	PublicationStatus *string    `json:"publication_status"`
	Summary           *string    `json:"summary"`
	StartYear         *int       `json:"start_year"`
	StartMonth        *int       `json:"start_month"`
	StartDay          *int       `json:"start_day"`
	Author            *string    `json:"author"`
	Art               *string    `json:"art"`
	WebLink           *string    `json:"web_link"`
	MetadataUpdatedAt *time.Time `json:"metadata_updated_at"`

	CreatedAt *time.Time `json:"created_at"`
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

type MangaPost struct {
	Name         string `json:"name"`
	ProviderName string `json:"provider"`
	TimeRule     string `json:"time_rule"`
	Url          string `json:"url"`
	MangaPath    string `json:"manga_path"`
}
