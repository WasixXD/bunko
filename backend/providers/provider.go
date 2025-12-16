package providers

import (
	"bunko/backend/core"
	"bunko/backend/providers/mangadex"
	"bunko/backend/providers/mangapill"
)

type Provider interface {
	// Function to return the first 3 links to where to find the manga
	Search(manga_name string) ([]string, error)

	// Function to get all Chapters from certain Provider
	GetAllChapters() ([]core.Chapter, error)

	// Download all chapter images
	DownloadChapter(url string) error
}

type ProviderFactory struct {
	providers map[string]Provider
}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: map[string]Provider{
			"mangadex":  &mangadex.Mangadex{},
			"mangapill": &mangapill.Mangapill{},
		},
	}
}

func (p *ProviderFactory) Get(name string) Provider {
	return p.providers[name]
}
