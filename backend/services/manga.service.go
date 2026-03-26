package services

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

type MangaService struct {
	db *sqlx.DB
}

func (m *MangaService) AddManga(manga structs.MangaPost) (int, error) {
	manga.MangaPath = normalizeFilesystemPath(manga.MangaPath)

	validation, err := m.ValidatePath(manga.MangaPath)
	if err != nil {
		return 0, err
	}
	if !validation.Valid {
		return 0, fmt.Errorf(validation.Message)
	}

	return db.AddMangaToDB(m.db, manga)
}

func (m *MangaService) GetAll() ([]structs.Manga, error) {
	return db.GetAllMangas(m.db)
}

func (m *MangaService) GetById(id string) (structs.Manga, error) {
	return db.GetMangaById(m.db, id)
}

func (m *MangaService) DeleteById(id string) (int, error) {
	return db.DeleteMangaById(m.db, id)
}

func (m *MangaService) AddTimeRule(time_rule, manga_id string) error {
	return db.AddTimeRule(m.db, time_rule, manga_id)
}

func (m *MangaService) ValidatePath(path string) (structs.PathValidationResponse, error) {
	cleanPath := normalizeFilesystemPath(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return structs.PathValidationResponse{
				Path:     cleanPath,
				Valid:    false,
				Exists:   false,
				CanWrite: false,
				Message:  "Path does not exist.",
			}, nil
		}
		return structs.PathValidationResponse{}, fmt.Errorf("could not inspect path: %w", err)
	}

	if !info.IsDir() {
		return structs.PathValidationResponse{
			Path:     cleanPath,
			Valid:    false,
			Exists:   true,
			CanWrite: false,
			Message:  "Path exists but is not a directory.",
		}, nil
	}

	testDir, err := os.MkdirTemp(cleanPath, ".bunko-write-test-*")
	if err != nil {
		return structs.PathValidationResponse{
			Path:     cleanPath,
			Valid:    false,
			Exists:   true,
			CanWrite: false,
			Message:  "Path is not writable or Bunko cannot create folders there.",
		}, nil
	}

	if removeErr := os.Remove(testDir); removeErr != nil {
		return structs.PathValidationResponse{}, fmt.Errorf("path is writable but cleanup failed: %w", removeErr)
	}

	return structs.PathValidationResponse{
		Path:     cleanPath,
		Valid:    true,
		Exists:   true,
		CanWrite: true,
		Message:  "Path is valid and writable.",
	}, nil
}

func (m *MangaService) SuggestPaths(path string) (structs.PathSuggestionResponse, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		trimmed = "."
	}

	expandedPath, homeDir := expandUserPath(trimmed)
	cleanPath := filepath.Clean(expandedPath)
	searchDir := cleanPath
	prefix := ""

	if !strings.HasSuffix(expandedPath, string(os.PathSeparator)) {
		searchDir = filepath.Dir(cleanPath)
		if searchDir == "." && filepath.IsAbs(cleanPath) {
			searchDir = string(os.PathSeparator)
		}
		prefix = filepath.Base(cleanPath)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return structs.PathSuggestionResponse{
				BasePath:    searchDir,
				Suggestions: []structs.PathSuggestion{},
			}, nil
		}
		return structs.PathSuggestionResponse{}, fmt.Errorf("could not list path suggestions: %w", err)
	}

	suggestions := make([]structs.PathSuggestion, 0)
	lowerPrefix := strings.ToLower(prefix)

	if trimmed == "." || trimmed == "./" || trimmed == "" {
		suggestions = append(suggestions, structs.PathSuggestion{
			Path:  ".",
			Label: "./",
		})
	}

	if homeDir != "" && (trimmed == "~" || trimmed == "~/" || trimmed == "" || strings.HasPrefix(trimmed, "~")) {
		suggestions = append(suggestions, structs.PathSuggestion{
			Path:  homeDir,
			Label: "~/",
		})
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if lowerPrefix != "" && !strings.HasPrefix(strings.ToLower(name), lowerPrefix) {
			continue
		}

		fullPath := filepath.Join(searchDir, name)
		displayPath := fullPath
		if homeDir != "" && (fullPath == homeDir || strings.HasPrefix(fullPath, homeDir+string(os.PathSeparator))) {
			displayPath = "~" + strings.TrimPrefix(fullPath, homeDir)
		} else if relPath, relErr := filepath.Rel(".", fullPath); relErr == nil && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			displayPath = "." + string(os.PathSeparator) + relPath
		}

		suggestions = append(suggestions, structs.PathSuggestion{
			Path:  displayPath,
			Label: name,
		})
	}

	suggestions = dedupePathSuggestions(suggestions)
	sort.SliceStable(suggestions, func(i, j int) bool {
		leftPriority := suggestionPriority(suggestions[i].Path)
		rightPriority := suggestionPriority(suggestions[j].Path)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return strings.ToLower(suggestions[i].Path) < strings.ToLower(suggestions[j].Path)
	})

	if len(suggestions) > 8 {
		suggestions = suggestions[:8]
	}

	return structs.PathSuggestionResponse{
		BasePath:    searchDir,
		Suggestions: suggestions,
	}, nil
}

func (m *MangaService) UpdateMetadata(id string) (structs.Manga, error) {
	manga, err := db.GetMangaById(m.db, id)
	if err != nil {
		return manga, err
	}

	metadata, err := structs.AnilistMetadataQuery(manga.Name)
	if err != nil {
		return manga, err
	}

	tx, err := m.db.Beginx()
	if err != nil {
		return manga, err
	}

	defer tx.Rollback()

	if err := db.AddMetadataToManga(tx, manga.MangaId, manga.Url, *metadata); err != nil {
		return manga, err
	}

	if manga.MangaPath != nil {
		if err := m.downloadCover(*manga.MangaPath, metadata.Data.Media.CoverImage.ExtraLarge); err != nil {
			return manga, err
		}
		if err := db.SetMangaCoverPath(tx, manga.MangaId, metadata.Data.Media.CoverImage.ExtraLarge); err != nil {
			return manga, err
		}
	}

	if err := tx.Commit(); err != nil {
		return manga, err
	}

	return db.GetMangaById(m.db, id)
}

func expandUserPath(path string) (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	if path == "~" && homeDir != "" {
		return homeDir, homeDir
	}

	if strings.HasPrefix(path, "~"+string(os.PathSeparator)) && homeDir != "" {
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/")), homeDir
	}

	return path, homeDir
}

func normalizeFilesystemPath(path string) string {
	expandedPath, _ := expandUserPath(strings.TrimSpace(path))
	return filepath.Clean(expandedPath)
}

func (m *MangaService) downloadCover(mangaPath, url string) error {
	absPath := filepath.Join(mangaPath, "cover.jpg")

	file, err := os.Create(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if _, err := file.Write(body); err != nil {
		return err
	}
	return nil
}

func IsDatabaseLockedError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "database is locked") ||
		errors.Is(err, os.ErrDeadlineExceeded)
}

func dedupePathSuggestions(suggestions []structs.PathSuggestion) []structs.PathSuggestion {
	seen := make(map[string]struct{}, len(suggestions))
	result := make([]structs.PathSuggestion, 0, len(suggestions))

	for _, suggestion := range suggestions {
		if _, ok := seen[suggestion.Path]; ok {
			continue
		}
		seen[suggestion.Path] = struct{}{}
		result = append(result, suggestion)
	}

	return result
}

func suggestionPriority(path string) int {
	switch {
	case path == "./" || path == ".":
		return 0
	case path == "~/" || path == "~":
		return 1
	case strings.HasPrefix(path, "./"):
		return 2
	case strings.HasPrefix(path, "~/"):
		return 3
	default:
		return 4
	}
}
