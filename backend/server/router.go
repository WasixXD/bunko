package server

import (
	"bunko/backend/services"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(serv *services.Services) *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	RegisterRoutes(r, serv)
	registerFrontend(r)
	return r
}

func registerFrontend(r *gin.Engine) {
	staticDir := StaticDir()
	indexPath := filepath.Join(staticDir, "index.html")

	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	r.StaticFS("/assets", gin.Dir(filepath.Join(staticDir, "assets"), false))
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Status(http.StatusNotFound)
			return
		}

		requestPath := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), "/")
		if requestPath != "." && requestPath != "" {
			staticPath := filepath.Join(staticDir, requestPath)
			if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
				c.File(staticPath)
				return
			}
		}

		c.File(indexPath)
	})
}
