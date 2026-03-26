package resolver

import (
	"bunko/backend/db"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func (r *Resolver) cleanupDeletedMangaFolders() error {
	paths, err := db.GetDeletedMangaPaths(r.Database)
	if err != nil {
		return err
	}

	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		if !isSafeMangaFolderPath(cleanPath) {
			log.Warn("[Resolver.cleanupDeletedMangaFolders] refusing to remove unsafe path", "path", cleanPath)
			continue
		}

		if err := os.RemoveAll(cleanPath); err != nil {
			return err
		}

		if err := db.DeleteQueuedMangaPath(r.Database, path); err != nil {
			return err
		}

		log.Info("[Resolver.cleanupDeletedMangaFolders] removed manga folder", "path", cleanPath)
	}

	return nil
}

func isSafeMangaFolderPath(path string) bool {
	if path == "" || path == "." || path == string(os.PathSeparator) {
		return false
	}

	base := filepath.Base(path)
	return base != "." && base != string(os.PathSeparator)
}
