package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"go-forum-project/auth-service/cmd/app"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:Qq1234567@localhost:5050/auth_db?sslmode=disable")

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	authApp := app.NewApp(db)
	authApp.Run()
}
