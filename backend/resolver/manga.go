package resolver

import (
	"bunko/backend/db"
	"bunko/backend/structs"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func (r *Resolver) checkNewManga() *structs.Manga {
	manga, err := db.GetNextPendingManga(r.Database)
	if err != nil {
		log.Error("[Resolver.checkNewManga] failed to query pending manga", "error", err)
		return nil
	}
	if manga == nil {
		return nil
	}

	log.Info("[Resolver] Found a new manga to download", "manga_id", manga.MangaId)
	return manga
}

func (r *Resolver) findChapters(mangaID int) ([]structs.Chapter, error) {
	providerName, url, err := db.GetMangaSource(r.Database, mangaID)
	if err != nil {
		log.Error("[Resolver.findChapters()] got error", "error", err)
		return nil, err
	}

	provider := r.Provider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("provider %q not found", providerName)
	}

	return provider.GetAllChapters(url)
}

func (r *Resolver) downloadCover(mangaID int, mangaPath, url string) error {
	absPath := fmt.Sprintf("%s/cover.jpg", mangaPath)

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

	b, _ := io.ReadAll(res.Body)
	if _, err = file.Write(b); err != nil {
		return err
	}

	if err := db.SetMangaCoverPathDB(r.Database, mangaID, url); err != nil {
		return err
	}
	return nil
}

func (r *Resolver) notifyWorkers() {
	r.Downloaders.Mutex.Lock()
	r.Downloaders.Cond.Broadcast()
	r.Downloaders.Mutex.Unlock()
}

func (r *Resolver) persistMangaData(
	manga *structs.Manga,
	metadata *structs.AnilistMetadataResponse,
	chapters []structs.Chapter,
) error {
	tx, err := r.Database.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = db.AddMetadataToManga(tx, manga.MangaId, manga.Url, *metadata); err != nil {
		return err
	}

	if err = db.AddChaptersToQueue(tx, manga.MangaId, *manga.MangaPath, chapters); err != nil {
		return err
	}

	if err = db.SetMangaStatusTx(tx, manga.MangaId, "downloading"); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Resolver) prepareFilesystem(manga *structs.Manga) error {
	dir := filepath.Dir(*manga.MangaPath + "/")
	return os.MkdirAll(dir, 0755)
}

func (r *Resolver) processNextManga() error {
	manga := r.checkNewManga()
	if manga == nil {
		return nil
	}

	if err := r.prepareFilesystem(manga); err != nil {
		return err
	}

	metadata, err := structs.AnilistMetadataQuery(manga.Name)
	if err != nil {
		return err
	}

	if err := r.downloadCover(manga.MangaId, *manga.MangaPath, metadata.Data.Media.CoverImage.ExtraLarge); err != nil {
		return err
	}

	chapters, err := r.findChapters(manga.MangaId)
	if err != nil {
		return err
	}

	if err := r.persistMangaData(manga, metadata, chapters); err != nil {
		return err
	}

	r.notifyWorkers()
	return nil
}
