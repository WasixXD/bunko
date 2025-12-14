package main

import (
	"bunko/backend/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return r
}

func main() {

	_, err := db.InitDb("bunko.db")

	if err != nil {
		log.Fatal(err)
	}

	r := setupRouter()

	_ = r.Run(":3000")
}
