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
	database, err := db.InitDb(server.DatabasePath())

	if err != nil {
		log.Fatal(err)
	}

	downloaders := downloader.NewDownloaderBy(50, database)
	resolver := resolver.NewResolver(server.TWO_SECONDS, database, downloaders)

	go resolver.Work()

	serv := services.NewServices(database, downloaders)
	r := server.SetupRouter(serv)

	fmt.Printf("Running Bunko %s on http://localhost%s\n", server.Version, server.ListenAddr())
	_ = r.Run(server.ListenAddr())
}
