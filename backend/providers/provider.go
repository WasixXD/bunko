package providers

import (
	"bunko/backend/core"
	"bunko/backend/providers/mangapill"

	"github.com/charmbracelet/log"
)

type Provider interface {
	// Function to return the first 3 links to where to find the manga
	Search(manga_name string) ([]core.MangaProps, error)

	// Function to get all Chapters from certain Provider
	GetAllChapters(url string) ([]core.Chapter, error)

	// Download all chapter images
	DownloadChapter(url, path, name string) error
}

type ProviderFactory struct {
	providers map[string]Provider
}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: map[string]Provider{
			// "mangadex":  &mangadex.Mangadex{},
			"mangapill": &mangapill.Mangapill{},
		},
	}
}

func (p *ProviderFactory) Get(name string) Provider {
	return p.providers[name]
}

func (p *ProviderFactory) FullSearch(s string) map[string][]core.MangaProps {

	response := make(map[string][]core.MangaProps)
	for key, value := range p.providers {
		props, err := value.Search(s)
		if err != nil {
			log.Error(err)
		}

		response[key] = props
	}

	return response
}
