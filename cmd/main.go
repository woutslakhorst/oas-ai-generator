package main

import (
	"log"

	"github.com/woutslakhorst/oas-ai-generator/internal/db"
	"github.com/woutslakhorst/oas-ai-generator/internal/server"
)

func main() {
	database := db.New()
	defer database.Close()

	r := server.New(database)

	log.Println("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
