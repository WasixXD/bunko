package main

import (
	"bunko/backend/db"
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

	resolver := resolver.NewResolver(server.TWO_SECONDS, database)

	go resolver.Work()

	serv := services.NewServices(database)
	r := server.SetupRouter(serv)

	fmt.Println("Running on http://localhost:3000")
	_ = r.Run(":3000")
}
