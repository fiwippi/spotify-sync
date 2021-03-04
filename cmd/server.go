package main

import (
	"github.com/fiwippi/spotify-sync/pkg/server"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
	"time"
)

var refresh time.Duration
var id, secret, redirect, serverKey, adminKey, port string

func init() {
	// Load the env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	// Get the spotify client variables and generate the url
	id = os.Getenv("SPOTIFY_ID")
	secret = os.Getenv("SPOTIFY_SECRET")
	redirect = os.Getenv("DOMAIN") + "/spotify-callback"
	if !strings.HasPrefix(redirect, "http://") {
		redirect = "http://" + redirect
	}

	// Get the server key and admin key and server port from the env
	serverKey = os.Getenv("SERVER_KEY")
	adminKey = os.Getenv("ADMIN_KEY")
	port = os.Getenv("PORT")
	refresh, _ = time.ParseDuration(os.Getenv("SYNC_REFRESH") + "s")
}

func main() {
	err := server.Run(serverKey, adminKey, id, secret, redirect, port, refresh)
	if err != nil {
		log.Fatal(err)
	}
}
