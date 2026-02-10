package structs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type AnilistQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type AnilistMetadataResponse struct {
	Data struct {
		Media struct {
			ID    int `json:"id"`
			Title struct {
				Native string `json:"native"`
			} `json:"title"`
			Status      string `json:"status"`
			Description string `json:"description"`
			StartDate   struct {
				Day   int `json:"day"`
				Month int `json:"month"`
				Year  int `json:"year"`
			} `json:"startDate"`
			Format string   `json:"format"`
			Genres []string `json:"genres"`
			Tags   []struct {
				Name string `json:"name"`
			} `json:"tags"`
			CoverImage struct {
				ExtraLarge string `json:"extraLarge"`
				Large      string `json:"large"`
				Medium     string `json:"medium"`
			} `json:"coverImage"`
			Staff struct {
				Edges []struct {
					Role string `json:"role"`
					Node struct {
						Name struct {
							Full string `json:"full"`
						} `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"staff"`
		} `json:"Media"`
	} `json:"data"`
}

func AnilistMetadataQuery(title string) (*AnilistMetadataResponse, error) {

	query := `
    query ($search: String) {
        Media(type: MANGA, search: $search) {
            id
            title {
                native
            }
            status
            description(asHtml: false)
            startDate {
                day
                month
                year
            }
            format
            genres
            tags {
                name
            }
			coverImage {
				extraLarge
				large
				medium
			}
            staff {
                edges {
                    role
                    node {
                        name {
                            full
                        }
                    }
                }
            }
        }
    }`
	payload := AnilistQuery{
		Query: query,
		Variables: map[string]interface{}{
			"search": title,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		"https://graphql.anilist.co",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result AnilistMetadataResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
