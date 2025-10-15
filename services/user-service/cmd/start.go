package cmd

import (
	"user-service/internal/app"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the User Service server",
	Long: `Start the User Service HTTP server dengan graceful shutdown.
Server akan berjalan di port yang didefinisikan di environment variable APP_PORT.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Override config dengan flags jika disediakan
		if port, _ := cmd.Flags().GetString("port"); port != "" {
			viper.Set("APP_PORT", port)
		}
		if env, _ := cmd.Flags().GetString("env"); env != "" {
			viper.Set("APP_ENV", env)
		}

		// Jalankan server
		app.RunServer()
	},
}

func init() {
	// Local flags untuk start command
	startCmd.Flags().StringP("port", "p", "", "Port untuk menjalankan server (override APP_PORT)")
	startCmd.Flags().StringP("env", "e", "", "Environment (override APP_ENV)")
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
