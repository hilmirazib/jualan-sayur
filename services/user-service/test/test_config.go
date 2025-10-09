package main

import (
	"log"

	"user-service/config"
)

func main() {
	// Test config loading
	cfg := config.NewConfig()

	log.Println("=== CONFIG TEST ===")
	log.Printf("App Port: %s", cfg.App.AppPort)
	log.Printf("App Env: %s", cfg.App.AppEnv)
	log.Printf("JWT Secret: %s", cfg.App.JwtSecretKey)
	log.Printf("JWT Issuer: %s", cfg.App.JwtIssuer)

	log.Println("\n=== DATABASE CONFIG ===")
	log.Printf("Host: %s", cfg.PsqlDB.Host)
	log.Printf("Port: %s", cfg.PsqlDB.Port)
	log.Printf("User: %s", cfg.PsqlDB.User)
	log.Printf("Database: %s", cfg.PsqlDB.DBName)
	log.Printf("Max Open: %d", cfg.PsqlDB.DBMaxOpen)
	log.Printf("Max Idle: %d", cfg.PsqlDB.DBMaxIdle)

	// Test database connection
	log.Println("\n=== DATABASE CONNECTION TEST ===")
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	log.Println("✅ Database connection successful!")
	log.Printf("Database object: %v", db.DB)

	// Test if we can ping the database
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	log.Println("✅ Database ping successful!")
	log.Println("✅ All tests passed! Config and database setup is working correctly.")
}
