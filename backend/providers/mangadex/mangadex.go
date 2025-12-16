package mangadex

import "bunko/backend/core"

type Mangadex struct {
}

func (m *Mangadex) Search(manga_name string) ([]string, error) {
	panic("TODO")
}

func (m *Mangadex) GetAllChapters() ([]core.Chapter, error) {
	panic("TODO")
}

func (m *Mangadex) DownloadChapter(url string) error {
	panic("TODO")
}
