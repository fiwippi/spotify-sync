package cmd

import (
	"github.com/fiwippi/spotify-sync/pkg/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

// TODO ensure wss
// TODO send no active device message to client instead of host
// TODO unpause clients once host unpauses
// TODO better errors for abnormal websocket connection

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Runs the client",
	Long:  `Runs the client to connect to the spotify server in the terminal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return client.NewClient().Run()
	},
}
