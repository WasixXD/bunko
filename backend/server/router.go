package server

import (
	"bunko/backend/services"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	BUNKO_VERSION = "1.0.0"
	BUNKO_DATABSE = "bunko.db"
	TWO_SECONDS   = time.Duration(2000 * time.Millisecond)
)

func SetupRouter(serv *services.Services) *gin.Engine {
	r := gin.Default()
	RegisterRoutes(r, serv)
	return r
}
