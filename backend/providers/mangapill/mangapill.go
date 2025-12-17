package mangapill

import (
	"bunko/backend/core"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const MANGA_PILL_DEFAULT_URL = "https://mangapill.com"

type Mangapill struct {
}

func (m *Mangapill) Search(manga_name string) ([]core.MangaProps, error) {
	url := fmt.Sprintf("%s/quick-search?q=%s", MANGA_PILL_DEFAULT_URL, manga_name)
	links := []core.MangaProps{}

	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return nil, err
	}

	doc.Find("a.bg-card").Each(func(i int, s *goquery.Selection) {

		if i >= 3 {
			return
		}

		href, _ := s.Attr("href")
		fullUrl := fmt.Sprintf("%s%s", MANGA_PILL_DEFAULT_URL, href)

		meta := s.Find("div.flex.flex-wrap > div")

		// TODO: Should download the cover image? Cache?
		img, _ := s.Find("img.object-cover").Attr("src")

		prop := core.MangaProps{
			Name:      s.Find("div.font-black").Text(),
			Year:      meta.Eq(1).Text(),
			Status:    meta.Eq(2).Text(),
			Url:       fullUrl,
			CoverPath: img,
		}

		links = append(links, prop)
	})

	return links, nil
}

func (m *Mangapill) GetAllChapters(url string) ([]core.Chapter, error) {
	return nil, nil
}

func (m *Mangapill) DownloadChapter(url string) error {
	return nil
}
