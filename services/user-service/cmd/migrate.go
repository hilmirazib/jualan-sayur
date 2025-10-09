package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	dir := flag.String("dir", "database/migrations", "migrations directory")
	dsn := flag.String("dsn", "", "database URL (overrides .env)")
	cmd := flag.String("cmd", "up", "migration command: up/down")
	flag.Parse()

	var databaseURL string
	if *dsn != "" {
		databaseURL = *dsn
	} else {
		host := os.Getenv("DATABASE_HOST")
		port := os.Getenv("DATABASE_PORT")
		user := os.Getenv("DATABASE_USER")
		pass := os.Getenv("DATABASE_PASSWORD")
		db := os.Getenv("DATABASE_NAME")
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	}

	m, err := migrate.New("file://"+*dir, databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize migrate: %v", err)
	}

	switch *cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up failed: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
	default:
		log.Fatalf("unknown cmd: %s", *cmd)
	}
}
