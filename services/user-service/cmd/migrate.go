package cmd

import (
	"fmt"
	"log"
	"user-service/database/seeds"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration tools",
	Long: `Tools untuk mengelola database migration dan seeding.

Subcommands:
- up: Jalankan semua migration yang belum dijalankan
- down: Rollback migration terakhir
- seed: Jalankan database seeding`,
}

// migrateUpCmd represents the migrate up command
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run database migrations",
	Long:  `Menjalankan semua migration yang belum dijalankan dan kemudian menjalankan seeding.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("up")
	},
}

// migrateDownCmd represents the migrate down command
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback database migrations",
	Long:  `Rollback migration terakhir yang telah dijalankan.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("down")
	},
}

// migrateSeedCmd represents the migrate seed command
var migrateSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeding",
	Long:  `Menjalankan database seeding untuk mengisi data awal.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSeeding()
	},
}

func init() {
	// Add subcommands ke migrate
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateSeedCmd)

	// Flags untuk migrate command
	migrateCmd.PersistentFlags().String("dir", "database/migrations", "migration directory")
	migrateCmd.PersistentFlags().String("dsn", "", "database URL (overrides .env)")
}

func runMigrations(cmd string) {
	dir, _ := migrateCmd.Flags().GetString("dir")
	dsn, _ := migrateCmd.Flags().GetString("dsn")

	var databaseURL string
	if dsn != "" {
		databaseURL = dsn
	} else {
		host := viper.GetString("DATABASE_HOST")
		port := viper.GetString("DATABASE_PORT")
		user := viper.GetString("DATABASE_USER")
		pass := viper.GetString("DATABASE_PASSWORD")
		db := viper.GetString("DATABASE_NAME")
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	}

	m, err := migrate.New("file://"+dir, databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize migrate: %v", err)
	}

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up failed: %v", err)
		}
		log.Println("✅ Migration completed successfully")

		// Run seeds after migration
		log.Println("🌱 Running database seeds...")
		runSeeding()
	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("migrate down failed: %v", err)
		}
		log.Println("✅ Migration rollback completed successfully")
	}
}

func runSeeding() {
	// Build database URL
	host := viper.GetString("DATABASE_HOST")
	port := viper.GetString("DATABASE_PORT")
	user := viper.GetString("DATABASE_USER")
	pass := viper.GetString("DATABASE_PASSWORD")
	db := viper.GetString("DATABASE_NAME")
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)

	// Initialize GORM connection for seeding
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database for seeding: %v", err)
	}

	// Run seeds
	seeds.SeedRole(gormDB)
	seeds.SeedUsers(gormDB)

	log.Println("✅ Seeding completed successfully")
}
