package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Accepts requests from a client and upgrades the connection to a
// shared connection. The server and client perform an initial
// handshake where the client is sent an authentication URL and is
// given 5 minutes to sign in. If approved then begin serving the client
func processIncomingUserWebsocket(c *gin.Context) {
	// Create a new user
	u := user{w: c.Writer, r: c.Request}

	// Upgrade the user's connection to a shared connection
	err := u.upgrade()
	if err != nil {
		Log.Debug().Err(err).Msg("Cannot upgrade user to websocket connection")
		return
	}

	// Perform the handshake to authenticate the user's spotify connection
	err = u.handshake()
	if err != nil {
		Log.Debug().Err(err).Msg("Cannot perform handshake with user")
		u.disconnect()
		return
	}

	// Generate spotify user data
	u.spotifyData, err = u.spotifyClient.CurrentUser()
	if err != nil {
		Log.Debug().Err(err).Str("Username", u.name).Msg("Cannot get user's spotify data")
		u.disconnect()
		return
	}

	// Serve the user (in a new goroutine)
	Log.Info().Str("Username", u.name).Str("Spotify Name", u.spotifyData.DisplayName).Msg("Serving user")
	go u.readPump()
}

// When user's use the authentication url they are redirected to this
// route where a spotify client to control their playback is created
func spotifyCallback (c *gin.Context) {
	// If state is incorrect then return 404
	st := c.Query("state")
	if !ids.Has(st) {
		Log.Debug().Str("state", st).Msg("State mismatch")
		http.Error(c.Writer, "State mismatch", http.StatusUnauthorized)
		return
	}

	// Authenticate the token
	token, err := auth.Token(st, c.Request)
	if err != nil {
		Log.Debug().Str("state", st).Msg("Couldn't retrieve token")
		http.Error(c.Writer, "Couldn't get token", http.StatusUnauthorized)
		return
	}
	tokenChans[st] <- token

	// Create a client using the specified token
	clientChans[st] <- auth.NewClient(token)
	Log.Debug().Str("state", st).Msg("Couldn't create spotify client")

	// Send a response back
	fmt.Fprintf(c.Writer, "Return to the client")
}
