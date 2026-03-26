package main

import (
	"bunko/backend/db"
	"bunko/backend/downloader"
	"bunko/backend/resolver"
	"bunko/backend/server"
	"bunko/backend/services"
	"fmt"
	"log"
)

func main() {

	database, err := db.InitDb(server.BUNKO_DATABSE)

	if err != nil {
		log.Fatal(err)
	}

	downloaders := downloader.NewDownloaderBy(50, database)
	resolver := resolver.NewResolver(server.TWO_SECONDS, database, downloaders)

	go resolver.Work()

	serv := services.NewServices(database, downloaders)
	r := server.SetupRouter(serv)

	fmt.Println("Running on http://localhost:3000")
	_ = r.Run(":3000")
}
