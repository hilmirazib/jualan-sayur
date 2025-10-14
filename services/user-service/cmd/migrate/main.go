package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"user-service/database/seeds"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
		log.Println("Migration completed successfully")

		// Run seeds after migration
		log.Println("Running database seeds...")
		runSeeds(databaseURL)
	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
	case "seed":
		runSeeds(databaseURL)
	default:
		log.Fatalf("unknown cmd: %s", *cmd)
	}
}

func runSeeds(databaseURL string) {
	// Initialize GORM connection for seeding
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database for seeding: %v", err)
	}

	// Run seeds
	seeds.SeedRole(db)
	seeds.SeedUsers(db)

	log.Println("Seeding completed successfully")
}
