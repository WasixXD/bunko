package main

import (
	"bunko/backend/db"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	BUNKO_VERSION = "1.0.0"
	BUNKO_DATABSE = "bunko.db"
)

func setupRouter() *gin.Engine {

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("Running Bunko %s", BUNKO_VERSION))
	})

	return r
}

func main() {

	_, err := db.InitDb(BUNKO_DATABSE)

	if err != nil {
		log.Fatal(err)
	}

	r := setupRouter()

	fmt.Println("Running on http://localhost:3000")
	_ = r.Run(":3000")
}
