package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/anvh2/futures-trading/internal/config"
	server "github.com/anvh2/futures-trading/internal/servers"
	"github.com/spf13/viper"
)

var cfg config.Config

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start futures-trading service",
	Long:  "Start futures-trading service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.Unmarshal(&cfg); err != nil {
			log.Fatalf("Unable to decode into struct: %v", err)
		}

		server := server.New(cfg)
		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
