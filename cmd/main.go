package main

import (
	"api-proxy/internal/api"
	"api-proxy/internal/config"
	"api-proxy/internal/db"
	"log"
)

func main() {
	appConfig, err := config.LoadConfig("config.yml")

	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Connect(appConfig.DB)

	if err != nil {
		log.Fatalf("Unable to connect to DB: %v\n", err)
	}

	if migrationErr := db.RunMigration(database); migrationErr != nil {
		log.Fatalf("Unable to run migrations: %v\n", migrationErr)
	}

	server := api.NewServer(appConfig, database)

	log.Printf("Listening on port %v", server.Port())
	log.Fatalln(server.Start())
}
