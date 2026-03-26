package structs

type ComicInfo struct {
	Title           string `xml:"Title" db:"title"` // chapter title
	LocalizedSeries string `xml:"LocalizedSeries" db:"localized_series"`
	Series          string `xml:"Series" db:"series"`
	// 0 if ended otherwise is ongoing
	PublicationStatus int    `xml:"Count" db:"publication_status"`
	Summary           string `xml:"Summary" db:"summary"`
	Writer            string `xml:"Writer" db:"writer"`
	Year              int    `xml:"Year" db:"year"`
	Month             int    `xml:"Month" db:"month"`
	Day               int    `xml:"Day" db:"day"`
	Web               string `xml:"Web" db:"web"`
	PageCount         int    `xml:"PageCount" db:"page_count"`
}
