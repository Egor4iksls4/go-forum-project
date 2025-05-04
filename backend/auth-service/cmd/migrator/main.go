package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var (
		action    string
		steps     int
		dbConnStr string
	)

	flag.StringVar(&action, "action", "up", "Миграция: up, down, force, version")
	flag.IntVar(&steps, "steps", 0, "Количество шагов (для up/down)")
	flag.StringVar(&dbConnStr, "db", "postgres://postgres:Qq1234567@localhost:5050/auth_db?sslmode=disable",
		"Строка подключения к БД")
	flag.Parse()

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://C:/Users/Egor/Desktop/go-forum/backend/auth-service/migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	switch action {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
	case "force":
		err = m.Force(steps)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n\n", version, dirty)
		return
	default:
		log.Fatalf("Неизвестное действие: %s", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	fmt.Println("Миграция успешно выполнена")
}
