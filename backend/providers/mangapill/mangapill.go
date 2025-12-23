package mangapill

import (
	"bunko/backend/core"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const MANGA_PILL_DEFAULT_URL = "https://mangapill.com"
const MANGA_PILL = "mangapill"

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
	chapters := []core.Chapter{}

	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return nil, err
	}

	doc.Find("#chapters a").Each(func(i int, s *goquery.Selection) {

		chapterUrl, _ := s.Attr("href")
		title, _ := s.Attr("title")

		fullUrl := fmt.Sprintf("%s%s", MANGA_PILL_DEFAULT_URL, chapterUrl)

		chapters = append(chapters, core.Chapter{
			Url:      fullUrl,
			Name:     strings.TrimSpace(title),
			Provider: MANGA_PILL,
		})
	})

	return chapters, nil
}

func (m *Mangapill) DownloadChapter(download_url, path_to_download, name string) error {

	res, err := http.Get(download_url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return err
	}

	doc.Find("chapter-page img").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("data-src")

		req, err := http.NewRequest(http.MethodGet, link, nil)

		if err != nil {
			return
		}

		req.Header.Set("Referer", MANGA_PILL_DEFAULT_URL)
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			return
		}

		if res.StatusCode != 200 {
			fmt.Println("Got differente status code", res.StatusCode)
			return
		}

		b, _ := io.ReadAll(res.Body)

		u, _ := url.Parse(link)
		last := path.Base(u.Path)

		// ./manga_path/manga_name/chapter_name/image
		absPath := fmt.Sprintf("%s/%s", path_to_download, last)

		if err := os.WriteFile(absPath, b, 0644); err != nil {
			fmt.Println(err)
			return
		}

	})
	return nil
}
