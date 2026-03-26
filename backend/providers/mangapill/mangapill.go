package mangapill

import (
	"bunko/backend/providers/providerutils"
	"bunko/backend/structs"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const defaultURL = "https://mangapill.com"
const providerName = "mangapill"

type Mangapill struct {
}

func (m *Mangapill) Search(manga_name string) ([]structs.MangaProps, error) {
	url := fmt.Sprintf("%s/quick-search?q=%s", defaultURL, url.QueryEscape(manga_name))
	links := []structs.MangaProps{}

	doc, err := providerutils.FetchDocument(url, providerutils.RequestConfig{})
	if err != nil {
		return nil, err
	}

	doc.Find("a.bg-card").Each(func(i int, s *goquery.Selection) {
		if i >= 3 {
			return
		}

		href, _ := s.Attr("href")
		fullURL, err := providerutils.ResolveURL(defaultURL, href)
		if err != nil {
			return
		}

		meta := s.Find("div.flex.flex-wrap > div")

		img, _ := s.Find("img.object-cover").Attr("src")
		coverPath, err := providerutils.ResolveURL(defaultURL, img)
		if err != nil {
			coverPath = img
		}

		prop := structs.MangaProps{
			Name:      strings.TrimSpace(s.Find("div.font-black").Text()),
			Year:      strings.TrimSpace(meta.Eq(1).Text()),
			Status:    strings.TrimSpace(meta.Eq(2).Text()),
			Url:       fullURL,
			CoverPath: coverPath,
		}

		links = append(links, prop)
	})

	return links, nil
}

func (m *Mangapill) GetAllChapters(url string) ([]structs.Chapter, error) {
	chapters := []structs.Chapter{}

	doc, err := providerutils.FetchDocument(url, providerutils.RequestConfig{})
	if err != nil {
		return nil, err
	}

	doc.Find("#chapters a").Each(func(i int, s *goquery.Selection) {
		chapterURL, _ := s.Attr("href")
		title, _ := s.Attr("title")

		fullURL, err := providerutils.ResolveURL(defaultURL, chapterURL)
		if err != nil {
			return
		}

		chapters = append(chapters, structs.Chapter{
			Url:      fullURL,
			Name:     strings.TrimSpace(title),
			Provider: providerName,
		})
	})

	return chapters, nil
}

func (m *Mangapill) DownloadChapter(download_url, path_to_download, _ string) error {
	doc, err := providerutils.FetchDocument(download_url, providerutils.RequestConfig{})
	if err != nil {
		return err
	}

	imageURLs := make([]string, 0)
	doc.Find("chapter-page img").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("data-src")
		if !ok {
			return
		}

		fullURL, err := providerutils.ResolveURL(defaultURL, link)
		if err != nil {
			return
		}

		imageURLs = append(imageURLs, fullURL)
	})

	for _, imageURL := range imageURLs {
		if err := providerutils.SaveURLToDir(imageURL, path_to_download, providerutils.RequestConfig{
			Headers: map[string]string{
				"Referer": defaultURL,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
