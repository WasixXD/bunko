package resolver

import (
	"bunko/backend/downloader"
	"bunko/backend/providers"
	"bunko/backend/structs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type fakeProvider struct {
	chapters []structs.Chapter
}

func (f fakeProvider) Search(string) ([]structs.MangaProps, error) {
	return nil, nil
}

func (f fakeProvider) GetAllChapters(string) ([]structs.Chapter, error) {
	return f.chapters, nil
}

func (f fakeProvider) DownloadChapter(string, string, string) error {
	return nil
}

func TestDiffChapterEnqueuesOnlyMissingChapters(t *testing.T) {
	db := setupResolverTestDB(t)

	mangaPath := t.TempDir()
	_, err := db.Exec(`
		INSERT INTO mangas (manga_id, name, slug, status, provider, url, manga_path)
		VALUES (1, 'Test Manga', 'test_manga', 'completed', 'fake', 'https://example.com/manga', ?)
	`, mangaPath)
	if err != nil {
		t.Fatalf("insert manga: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO download_queue (manga_id, name, url, status, provider, path_to_download)
		VALUES
			(1, 'Chapter 1', 'https://example.com/ch1', 'completed', 'fake', ?),
			(1, 'Chapter 2', 'https://example.com/ch2', 'completed', 'fake', ?)
	`, filepath.Join(mangaPath, "Chapter 1"), filepath.Join(mangaPath, "Chapter 2"))
	if err != nil {
		t.Fatalf("insert queue rows: %v", err)
	}

	mutex := &sync.Mutex{}
	resolver := &Resolver{
		Database: db,
		Downloaders: &downloader.DownloaderLock{
			Mutex: mutex,
			Cond:  sync.NewCond(mutex),
		},
		Provider: func(string) providers.Provider {
			return fakeProvider{
				chapters: []structs.Chapter{
					{Name: "Chapter 1", Url: "https://example.com/ch1", Provider: "fake"},
					{Name: "Chapter 2", Url: "https://example.com/ch2", Provider: "fake"},
					{Name: "Chapter 3", Url: "https://example.com/ch3", Provider: "fake"},
				},
			}
		},
	}

	if err := resolver.diffChapter(1); err != nil {
		t.Fatalf("diffChapter: %v", err)
	}

	var rows []struct {
		Name   string `db:"name"`
		Status string `db:"status"`
		Path   string `db:"path_to_download"`
	}
	err = db.Select(&rows, `
		SELECT name, status, path_to_download
		FROM download_queue
		WHERE manga_id = 1
		ORDER BY name
	`)
	if err != nil {
		t.Fatalf("select queue: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 queue rows, got %d", len(rows))
	}

	if rows[2].Name != "Chapter 3" {
		t.Fatalf("expected missing chapter to be enqueued, got %#v", rows[2])
	}

	if rows[2].Status != "pending" {
		t.Fatalf("expected new chapter status pending, got %q", rows[2].Status)
	}

	expectedPath := filepath.Join(mangaPath, "Chapter 3")
	if rows[2].Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, rows[2].Path)
	}

	var mangaStatus string
	if err := db.Get(&mangaStatus, `SELECT status FROM mangas WHERE manga_id = 1`); err != nil {
		t.Fatalf("select manga status: %v", err)
	}

	if mangaStatus != "downloading" {
		t.Fatalf("expected manga status downloading, got %q", mangaStatus)
	}
}

func TestDiffChapterNoopWhenNothingIsMissing(t *testing.T) {
	db := setupResolverTestDB(t)

	mangaPath := t.TempDir()
	_, err := db.Exec(`
		INSERT INTO mangas (manga_id, name, slug, status, provider, url, manga_path)
		VALUES (1, 'Test Manga', 'test_manga', 'completed', 'fake', 'https://example.com/manga', ?)
	`, mangaPath)
	if err != nil {
		t.Fatalf("insert manga: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO download_queue (manga_id, name, url, status, provider, path_to_download)
		VALUES (1, 'Chapter 1', 'https://example.com/ch1', 'completed', 'fake', ?)
	`, filepath.Join(mangaPath, "Chapter 1"))
	if err != nil {
		t.Fatalf("insert queue row: %v", err)
	}

	mutex := &sync.Mutex{}
	resolver := &Resolver{
		Database: db,
		Downloaders: &downloader.DownloaderLock{
			Mutex: mutex,
			Cond:  sync.NewCond(mutex),
		},
		Provider: func(string) providers.Provider {
			return fakeProvider{
				chapters: []structs.Chapter{
					{Name: "Chapter 1", Url: "https://example.com/ch1", Provider: "fake"},
				},
			}
		},
	}

	if err := resolver.diffChapter(1); err != nil {
		t.Fatalf("diffChapter: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM download_queue WHERE manga_id = 1`); err != nil {
		t.Fatalf("count queue rows: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected queue to remain unchanged, got %d rows", count)
	}
}

func setupResolverTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db, err := sqlx.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	migrationPath := filepath.Join("..", "migrations", "1_create_table.up.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	if _, err := db.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
