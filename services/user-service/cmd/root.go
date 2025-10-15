package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sayur-api",
	Short: "Sayur API - Microservices untuk aplikasi jualan sayur",
	Long: `Sayur API adalah kumpulan microservices untuk aplikasi jualan sayur
yang dibangun dengan Go, Echo framework, dan PostgreSQL.

Microservices yang tersedia:
- User Service: Mengelola autentikasi dan data pengguna
- Product Service: Mengelola data produk sayur
- Order Service: Mengelola pesanan
- Payment Service: Mengelola pembayaran
- Notification Service: Mengelola notifikasi

Untuk informasi lebih lanjut, kunjungi: https://github.com/hilmirazib/jualan-sayur`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(configCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find current directory.
		viper.AddConfigPath(".")
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose, _ := rootCmd.Flags().GetBool("verbose"); verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}
