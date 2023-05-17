package cmd

import (
	"github.com/spf13/cobra"

	"github.com/anvh2/futures-trading/internal/server"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start futures-trading service",
	Long:  "Start futures-trading service",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := server.New()
		return server.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
