package main

import (
	"log"

	"github.com/example/petstore/internal/db"
	"github.com/example/petstore/internal/server"
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
