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

func (m *Mangapill) Search(manga_name string) ([]string, error) {
	url := fmt.Sprintf("%s/quick-search?q=%s", MANGA_PILL_DEFAULT_URL, manga_name)
	links := []string{}

	res, err := http.Get(url)

	fmt.Println(url)
	if err != nil {
		return links, err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return links, nil
	}

	doc.Find("a.bg-card").Each(func(i int, s *goquery.Selection) {

		if i >= 3 {
			return
		}

		href, _ := s.Attr("href")

		fullUrl := fmt.Sprintf("%s%s", MANGA_PILL_DEFAULT_URL, href)
		links = append(links, fullUrl)
	})

	return links, nil
}

func (m *Mangapill) GetAllChapters() ([]core.Chapter, error) {
	return nil, nil
}

func (m *Mangapill) DownloadChapter(url string) error {
	return nil
}
