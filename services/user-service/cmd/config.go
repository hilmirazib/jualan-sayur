package cmd

import (
	"fmt"
	"log"
	"user-service/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long: `Menampilkan konfigurasi yang sedang aktif dari environment variables
dan file konfigurasi yang digunakan.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig(cmd)
	},
}

func init() {
	// Flags untuk config command
	configCmd.Flags().Bool("validate-db", false, "validate database connection")
}

func showConfig(cmd *cobra.Command) {
	fmt.Println("🔧 Current Configuration")
	fmt.Println("========================")

	// App Configuration
	fmt.Println("\n📱 App Configuration:")
	fmt.Printf("  APP_PORT: %s\n", viper.GetString("APP_PORT"))
	fmt.Printf("  APP_ENV: %s\n", viper.GetString("APP_ENV"))
	fmt.Printf("  JWT_SECRET_KEY: %s\n", maskSecret(viper.GetString("JWT_SECRET_KEY")))
	fmt.Printf("  JWT_ISSUER: %s\n", viper.GetString("JWT_ISSUER"))

	// Database Configuration
	fmt.Println("\n🗄️  Database Configuration:")
	fmt.Printf("  DATABASE_HOST: %s\n", viper.GetString("DATABASE_HOST"))
	fmt.Printf("  DATABASE_PORT: %s\n", viper.GetString("DATABASE_PORT"))
	fmt.Printf("  DATABASE_USER: %s\n", viper.GetString("DATABASE_USER"))
	fmt.Printf("  DATABASE_PASSWORD: %s\n", maskSecret(viper.GetString("DATABASE_PASSWORD")))
	fmt.Printf("  DATABASE_NAME: %s\n", viper.GetString("DATABASE_NAME"))
	fmt.Printf("  DATABASE_MAX_OPEN_CONNECTION: %d\n", viper.GetInt("DATABASE_MAX_OPEN_CONNECTION"))
	fmt.Printf("  DATABASE_MAX_IDLE_CONNECTION: %d\n", viper.GetInt("DATABASE_MAX_IDLE_CONNECTION"))

	// Config file info
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		fmt.Printf("\n📄 Config file: %s\n", configFile)
	} else {
		fmt.Println("\n📄 Config file: Not found (using environment variables)")
	}

	// Validate database connection if requested
	if validateDB, _ := cmd.Flags().GetBool("validate-db"); validateDB {
		fmt.Println("\n🔍 Validating database connection...")
		validateDatabaseConnection()
	}
}

func maskSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:4] + "****"
}

func validateDatabaseConnection() {
	cfg := config.NewConfig()

	// Try to connect to database
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Printf("❌ Database connection failed: %v", err)
		return
	}

	// Test the connection
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("❌ Failed to get database instance: %v", err)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		log.Printf("❌ Database ping failed: %v", err)
		return
	}

	log.Println("✅ Database connection successful")
}
