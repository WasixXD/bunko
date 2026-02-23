package server

import (
	"bunko/backend/providers"
	"bunko/backend/services"
	"bunko/backend/structs"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

func HandleMain(serv *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("Running Bunko %s", BUNKO_VERSION))
	}
}

func HandleAddManga(serv *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {

		var json structs.MangaPost

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		id, err := serv.Manga.AddManga(json)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusCreated, gin.H{"manga_id": id})
	}
}

func HandleQuickSearchManga(serv *services.Services) gin.HandlerFunc {
	factory := providers.NewProviderFactory()

	// TODO: Handle this into the service
	return func(c *gin.Context) {
		q := c.Query("q")
		mangas := factory.FullSearch(q)
		c.JSON(http.StatusOK, mangas)
	}
}

func HandleMangas(serv *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		mangas, err := serv.Manga.GetAll()

		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, mangas)
	}
}

func HandleMangasGet(serv *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("id")

		manga, err := serv.Manga.GetById(q)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, manga)
	}
}

func HandleQueue(serv *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		clientChannel := c.Writer.CloseNotify()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-clientChannel:
				log.Info("[Server] Client disconnected")
				return

			case <-ticker.C:

				jobs, err := serv.Queue.GetAll()

				if err != nil {
					log.Error(err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}

				jsonData, _ := json.Marshal(jobs)
				fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
				c.Writer.Flush()
			}
		}

	}
}

func RegisterRoutes(r *gin.Engine, serv *services.Services) {
	r.GET("/", HandleMain(serv))
	r.POST("/add/manga", HandleAddManga(serv))
	r.GET("/quick-search/manga", HandleQuickSearchManga(serv))
	r.GET("/mangas", HandleMangas(serv))
	r.GET("/mangas/get/", HandleMangasGet(serv))
	r.GET("/queue", HandleQueue(serv))
}
