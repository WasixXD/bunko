package core

type ComicInfo struct {
	Title           string `xml:"Title"` // chapter title
	LocalizedSeries string `xml:"LocalizedSeries"`
	Series          string `xml:"Series"`
	// 0 if ended otherwise is ongoing
	PublicationStatus int    `xml:"Count"`
	Summary           string `xml:"Summary"`
	Year              int    `xml:"Year"`
	Month             int    `xml:"Month"`
	Day               int    `xml:"Day"`
	Web               string `xml:"Web"`
	PageCount         int    `xml:"PageCount"`
}
