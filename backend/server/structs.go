package server

type MangaPost struct {
	Name         string `json:"name"`
	ProviderName string `json:"provider"`
	TimeRule     string `json:"time_rule"`
	Url          string `json:"url"`
}
