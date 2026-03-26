package resolver

import (
	"bunko/backend/db"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupDeletedMangaFoldersRemovesQueuedFolder(t *testing.T) {
	testDB := setupResolverTestDB(t)

	mangaRoot := t.TempDir()
	mangaFolder := filepath.Join(mangaRoot, "Test Manga")
	if err := os.MkdirAll(mangaFolder, 0755); err != nil {
		t.Fatalf("create manga folder: %v", err)
	}

	_, err := testDB.Exec(`
		INSERT INTO mangas (manga_id, name, slug, status, provider, url, manga_path)
		VALUES (1, 'Test Manga', 'test_manga', 'completed', 'fake', 'https://example.com/manga', ?)
	`, mangaFolder)
	if err != nil {
		t.Fatalf("insert manga: %v", err)
	}

	rows, err := db.DeleteMangaById(testDB, "1")
	if err != nil {
		t.Fatalf("delete manga: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected one deleted row, got %d", rows)
	}

	resolver := &Resolver{Database: testDB}
	if err := resolver.cleanupDeletedMangaFolders(); err != nil {
		t.Fatalf("cleanupDeletedMangaFolders: %v", err)
	}

	if _, err := os.Stat(mangaFolder); !os.IsNotExist(err) {
		t.Fatalf("expected manga folder to be deleted, stat err = %v", err)
	}

	queuedPaths, err := db.GetDeletedMangaPaths(testDB)
	if err != nil {
		t.Fatalf("get deleted manga paths: %v", err)
	}

	if len(queuedPaths) != 0 {
		t.Fatalf("expected cleanup queue to be empty, got %d entries", len(queuedPaths))
	}
}
