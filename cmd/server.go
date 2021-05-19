package cmd

import (
	"errors"
	"fmt"
	"github.com/fiwippi/spotify-sync/pkg/server"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var refresh time.Duration
var id, secret, redirect, serverKey, adminKey, port, envPath, ssl, mode, logLevel string

var validLogLevels = map[string]bool{
	"trace": true,
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"fatal": true,
	"panic": true,
}

func init() {
	serverCmd.Flags().StringVar(&ssl, "ssl", "", "is the server running behind ssl (default true)")
	serverCmd.Flags().StringVarP(&redirect, "domain", "d", "", "domain server is running on, e.g. spotify.site.net")
	serverCmd.Flags().StringVarP(&port, "port", "p", "", "port to host the server on (default 8096)")
	serverCmd.Flags().StringVarP(&envPath, "env-path", "e", ".env", "path to load env file from")
	serverCmd.Flags().StringVarP(&mode, "mode", "m", "", "runs the server from available modes \"debug\", \"release\" (default \"debug\")")
	serverCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "log level from \"trace\", \"debug\", \"info\", \"warn\", \"error\", \"fatal\", \"panic\" (default \"debug\")")
	serverCmd.Flags().DurationVarP(&refresh, "refresh-interval", "r", 0, "how often should the server attempt to sync the session (default 10s)")

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the server",
	Long:  `Runs the sync server which accepts clients`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flag args
		if ssl != "" && ssl != "false" && ssl != "true" {
			return errors.New("ssl must be \"true\" or \"false\" (case sensitive)")
		}

		// Load the env file
		godotenv.Load(envPath)

		// Get spotify secrets
		id = os.Getenv("SPOTIFY_ID")
		secret = os.Getenv("SPOTIFY_SECRET")

		// Get server secrets
		adminKey = os.Getenv("ADMIN_KEY")

		// Load the rest
		if redirect == "" {
			redirect = os.Getenv("DOMAIN") + "/spotify-callback"
		}

		if ssl == "" {
			if os.Getenv("USE_SSL") == "false" {
				redirect = "http://" + redirect
			} else {
				redirect = "https://" + redirect
			}
		} else if ssl == "false" {
			redirect = "http://" + redirect
		} else {
			redirect = "https://" + redirect
		}

		if port == "" {
			if os.Getenv("PORT") != "" {
				port = os.Getenv("PORT")
			} else {
				port = "8096"
			}
		}

		if refresh == 0 {
			if os.Getenv("SYNC_REFRESH") != "" {
				refresh, _ = time.ParseDuration(os.Getenv("SYNC_REFRESH") + "s")
			} else {
				refresh = 10 * time.Second
			}
		}

		if mode == "" {
			if os.Getenv("SERVER_MODE") != "" {
				mode = os.Getenv("SERVER_MODE")
			} else {
				mode = "debug"
			}
		}

		if logLevel == "" {
			if os.Getenv("SERVER_LOG_LEVEL") != "" {
				logLevel = os.Getenv("SERVER_LOG_LEVEL")
			} else {
				logLevel = "debug"
			}
		} else if found, _ := validLogLevels[logLevel]; !found {
			return errors.New("invalid log level, must be one of \"trace\", \"debug\", \"info\", \"warn\", \"error\", \"fatal\", \"panic\"")
		}

		fmt.Printf("Server running with config:\nredirect - %s\nport - %s\nrefresh - %s\nmode - %s\nlog level - %s\n", redirect, port, refresh, mode, logLevel)
		return server.Run(adminKey, id, secret, redirect, port, mode, logLevel, refresh)
	},
}
