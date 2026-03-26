package providerutils

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

// ResolveURL resolves relative provider links against a base provider URL.
func ResolveURL(baseURL, ref string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url %q: %w", baseURL, err)
	}

	target, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("parse ref url %q: %w", ref, err)
	}

	return base.ResolveReference(target).String(), nil
}

// URLFilename extracts the filename portion from a URL path.
func URLFilename(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse file url %q: %w", rawURL, err)
	}

	filename := path.Base(parsed.Path)
	if filename == "." || filename == "/" || filename == "" {
		return "", fmt.Errorf("url %q does not contain a filename", rawURL)
	}

	return filename, nil
}

// SaveURLToDir downloads a URL and saves it to destDir using the URL filename.
func SaveURLToDir(rawURL, destDir string, config RequestConfig) error {
	body, err := FetchBody(rawURL, config)
	if err != nil {
		return err
	}
	defer body.Close()

	filename, err := URLFilename(rawURL)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(destDir, filename)
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create %q: %w", targetPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return fmt.Errorf("write %q: %w", targetPath, err)
	}

	return nil
}
