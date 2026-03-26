package server

import (
	"os"
	"time"
)

var Version = "dev"

const TWO_SECONDS = time.Duration(2000 * time.Millisecond)

func DatabasePath() string {
	if value := os.Getenv("BUNKO_DATABASE"); value != "" {
		return value
	}

	return "bunko.db"
}

func StaticDir() string {
	if value := os.Getenv("BUNKO_STATIC_DIR"); value != "" {
		return value
	}

	return "./public/browser"
}

func DefaultMangaPath() string {
	if value := os.Getenv("BUNKO_MANGA_PATH"); value != "" {
		return value
	}

	return "./mangas"
}

func ListenAddr() string {
	if value := os.Getenv("BUNKO_LISTEN_ADDR"); value != "" {
		return value
	}

	return ":3000"
}
