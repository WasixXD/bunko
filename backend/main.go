package main

import (
	"bunko/backend/db"
	"bunko/backend/providers"
	"bunko/backend/resolver"
	"bunko/backend/server"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)


const (
	BUNKO_VERSION = "1.0.0"
	BUNKO_DATABSE = "bunko.db"
	TWO_SECONDS   = time.Duration(2000 * time.Millisecond)
)

// TODO: Set up this is another file
func setupRouter(database *sql.DB) *gin.Engine {
	factory := providers.NewProviderFactory()
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("Running Bunko %s", BUNKO_VERSION))
	})

	r.POST("/add/manga", func(c *gin.Context) {
		var json server.MangaPost

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		id, err := db.AddMangaToDB(database, json)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusCreated, gin.H{"manga_id": id})
	})

	r.GET("/quick-search/manga", func(c *gin.Context) {
		q := c.Query("q")

		mangas := factory.FullSearch(q)

		c.JSON(http.StatusOK, mangas)
	})

	return r
}

func main() {

	database, err := db.InitDb(BUNKO_DATABSE)

	if err != nil {
		log.Fatal(err)
	}

	resolver := resolver.NewResolver(TWO_SECONDS, database)

	go resolver.Work()

	r := setupRouter(database)

	fmt.Println("Running on http://localhost:3000")
	_ = r.Run(":3000")
}
