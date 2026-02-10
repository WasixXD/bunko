package mangadex

import "bunko/backend/structs"

type Mangadex struct {
}

func (m *Mangadex) Search(manga_name string) ([]string, error) {
	panic("TODO")
}

func (m *Mangadex) GetAllChapters() ([]structs.Chapter, error) {
	panic("TODO")
}

func (m *Mangadex) DownloadChapter(url string) error {
	panic("TODO")
}
