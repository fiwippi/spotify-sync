package server

import (
	"errors"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Whether the authenticator has been created, used in authGenerated()
// to ensure it exists before routes requiring it can be accessed i.e. /spotify-callback
var authCreated bool = false
// The spotify authenticator object used to create clients for each user
var auth spotify.Authenticator

// Generates the authenticator for the router
func generateAuth(id, secret, redirect string) error {
	// Fails if one of the variables is not set
	if len(id) == 0 || len(secret) == 0 || len(redirect) == 0 {
		return errors.New("Cannot setup router because one of the spotify config params was not set")
	}

	// Create the authenticator for the spotify session and generate its url
	auth = spotify.NewAuthenticator(redirect, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadPlaybackState)
	auth.SetAuthInfo(id, secret)
	authCreated = true
	return nil
}

// Checks if the channel used to send spotify clients generated in the
// SpotifyCallback to the user handshake where the user is created, is closed
func isSpotifyClientChanClosed(ch <-chan spotify.Client) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

// Checks if the channel used to send the spotify oauth2.Token generated in the
// SpotifyCallback to the user handshake where the user is created, is closed
func isSpotifyTokenChanClosed(ch <-chan *oauth2.Token) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
