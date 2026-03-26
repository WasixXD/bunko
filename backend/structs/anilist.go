package structs

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
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
						PrimaryOccupations []string `json:"primaryOccupations"`
						Name               struct {
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
						primaryOccupations
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

func (r AnilistMetadataResponse) Creators() (string, string) {
	authorSet := map[string]struct{}{}
	artSet := map[string]struct{}{}

	for _, edge := range r.Data.Media.Staff.Edges {
		name := strings.TrimSpace(edge.Node.Name.Full)
		if name == "" {
			continue
		}

		role := strings.ToLower(strings.TrimSpace(edge.Role))
		roleMarksStory := strings.Contains(role, "story") || strings.Contains(role, "writer")
		roleMarksArt := strings.Contains(role, "art") || strings.Contains(role, "artist") || strings.Contains(role, "illustrat")

		if roleMarksStory {
			authorSet[name] = struct{}{}
		}
		if roleMarksArt {
			artSet[name] = struct{}{}
		}

		if roleMarksStory || roleMarksArt {
			continue
		}

		for _, occupation := range edge.Node.PrimaryOccupations {
			occupation = strings.ToLower(strings.TrimSpace(occupation))
			switch {
			case strings.Contains(occupation, "writer"):
				authorSet[name] = struct{}{}
			case strings.Contains(occupation, "mangaka"), strings.Contains(occupation, "artist"), strings.Contains(occupation, "illustrat"):
				artSet[name] = struct{}{}
			}
		}
	}

	return joinCreatorNames(authorSet), joinCreatorNames(artSet)
}

func joinCreatorNames(values map[string]struct{}) string {
	names := make([]string, 0, len(values))
	for value := range values {
		names = append(names, value)
	}
	if len(names) == 0 {
		return ""
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
