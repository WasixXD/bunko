package providerutils

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const defaultUserAgent = "bunko/1.0 (+https://github.com/wasix/bunko)"

var defaultClient = &http.Client{
	Timeout: 30 * time.Second,
}

type RequestConfig struct {
	Headers map[string]string
}

// FetchDocument retrieves an HTML page and parses it into a goquery document.
func FetchDocument(rawURL string, config RequestConfig) (*goquery.Document, error) {
	body, err := FetchBody(rawURL, config)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("parse html %q: %w", rawURL, err)
	}

	return doc, nil
}

// FetchBody performs a GET request and returns the response body for successful responses.
func FetchBody(rawURL string, config RequestConfig) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request %q: %w", rawURL, err)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	res, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %q: %w", rawURL, err)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		defer res.Body.Close()
		return nil, fmt.Errorf("request %q: unexpected status %d", rawURL, res.StatusCode)
	}

	return res.Body, nil
}
