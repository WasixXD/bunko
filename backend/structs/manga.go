package structs

import (
	"strings"
	"time"
)

// This struct represents the main information from database mangas table
type Manga struct {
	MangaId int `json:"manga_id" db:"manga_id"`

	Name string `json:"name" db:"name"`
	Slug string `json:"slug" db:"slug"`

	// System status
	Status string `json:"status" db:"status"`

	Provider string `json:"provider" db:"provider"`
	Url      string `json:"url" db:"url"`

	CoverPath *string `json:"cover_path" db:"cover_path"`
	MangaPath *string `json:"manga_path" db:"manga_path"`

	LocalizedName     *string    `json:"localized_name" db:"localized_name"`
	PublicationStatus *string    `json:"publication_status" db:"publication_status"`
	Summary           *string    `json:"summary" db:"summary"`
	StartYear         *int       `json:"start_year" db:"start_year"`
	StartMonth        *int       `json:"start_month" db:"start_month"`
	StartDay          *int       `json:"start_day" db:"start_day"`
	Author            *string    `json:"author" db:"author"`
	WebLink           *string    `json:"web_link" db:"web_link"`
	MetadataUpdatedAt *time.Time `json:"metadata_updated_at" db:"metadata_updated_at"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
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

type PathValidationRequest struct {
	Path string `json:"path"`
}

type PathValidationResponse struct {
	Path     string `json:"path"`
	Valid    bool   `json:"valid"`
	Exists   bool   `json:"exists"`
	CanWrite bool   `json:"can_write"`
	Message  string `json:"message"`
}

type PathSuggestion struct {
	Path  string `json:"path"`
	Label string `json:"label"`
}

type PathSuggestionResponse struct {
	BasePath    string           `json:"base_path"`
	Suggestions []PathSuggestion `json:"suggestions"`
}
