package server

import (
	"errors"
	"fmt"
	ws "github.com/fiwippi/spotify-sync/pkg/shared"
	"github.com/zmb3/spotify"
	"log"
	"strings"
	"time"
)

// Code adapted from https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go

// How often in seconds to make calls the spotify api to ensure host and client are synced
var syncRefresh time.Duration

// Map which maps usernames hosting the session to their respective session
var sessions = make(map[string]*session)

// Session used to handleSync playback between a host and other clients
type session struct {
	clients    map[*user]bool // Registered clients.
	register   chan *user     // Register requests from the clients.
	unregister chan *user     // Unregister requests from clients.
	done       chan error     // Signals session to stop running (stops the handleChannels() function)
	broadcast  chan string    // Channel to receive string messages to send to other clients
	host       *user          // The user hosting the session
	quit       chan struct{}  // Channel to tell the session to stop synchronising (stops the handleSync() function)
}

// Initialiases a new session
func newSession(host *user) *session {
	s := &session{
		register:   make(chan *user),
		unregister: make(chan *user),
		done:       make(chan error),
		quit:       make(chan struct{}),
		broadcast:  make(chan string),
		clients:    make(map[*user]bool),
		host:       host,
	}
	s.clients[host] = true
	host.s = s

	return s
}

// Determines whether the unregister channel is closed,
// this is used when disconnecting the user
func unregisterChannelClosed(ch <-chan *user) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

// Sends a list of clients to all clients in the session
func (s *session) sendUserUpdate() error {
	users := s.getUsers()

	for client := range s.clients {
		// Send the user list
		msg := &ws.Message{
			Op:        "USERS",
			Args:      nil,
			Body:      users,
			Timestamp: ws.CurrentTime(),
		}

		err := client.WriteJSON(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Generates a string of all users within the session
func (s *session) getUsers() string {
	u := ""
	for client := range s.clients {
		u += client.name + ","
	}

	return strings.TrimSuffix(u, ",")
}

// Closes a session and deletes it from the session map
func (s *session) close() {
	sessions[s.host.name] = nil
	s.done <- errors.New("Closing session") // Stops the handleChannels() func

	// Notifies that the session is closed for all clients
	for client := range s.clients {
		client.sendInfo("Session (" + s.host.name + ") closed")
		client.clearUserList() // Tells the client no more users are in the session
	}

	// No more clients can be registered/unregistered
	close(s.register)
	close(s.unregister)
}

// Handles incoming/outgoing clients and
// broadcasting messages between clients.
func (s *session) handleChannels() {
	for {
		select {
		case _ = <-s.done:
			close(s.quit) // Stops the handleSync() func
			return
		case client := <-s.register:
			s.clients[client] = true
			_ = s.sendUserUpdate()
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
			}
			_ = s.sendUserUpdate()
		case text := <-s.broadcast:
			for client := range s.clients {
				err := client.sendMsg(text)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

// Syncs all clients in the session to have the
// same spotify playback as the host
func (s *session) handleSync() {
	ticker := time.NewTicker(syncRefresh)
	for {
		select {
		case <-ticker.C:
			// Get the host's spotify state
			hostState, err := s.host.spotifyClient.PlayerState()
			Log.Trace().Str("Username", s.host.name).Bool("Host", true).Str("State", fmt.Sprintf("%+v", hostState)).Msg("")

			// If there is an error, skip this iteration of the ticker
			if err != nil {
				Log.Warn().Str("Username", s.host.name).Bool("Host", true).Err(err).Msg("Spotify player state error")
				continue
			}

			// No active device
			if hostState.Device == (spotify.PlayerDevice{}) {
				s.host.sendInfo("You have no active device to play from...")
				continue
			}

			// Time to measure delay between checking the host and client progress to account for that
			startTime := time.Now()

			// Go through each client
			for client := range s.clients {
				// Avoid the host
				if client != s.host {
					// If the host is paused then pause the client
					if !hostState.Playing {
						err = client.spotifyClient.Pause()
						if err != nil {
							Log.Warn().Str("Username", client.name).Bool("Host", false).Err(err).Msg("Error pausing client")
						}
					} else {
						// Otherwise match the client to the player state of the host

						// Get the state of each client
						clientState, err := client.spotifyClient.PlayerState()
						Log.Trace().Str("Username", client.name).Bool("Host", false).Str("State", fmt.Sprintf("%+v", clientState)).Msg("")
						if err != nil {
							Log.Warn().Str("Username", client.name).Bool("Host", false).Err(err).Msg("Spotify player state error for")
							continue
						}

						// No active device
						if clientState.Device == (spotify.PlayerDevice{}) {
							s.host.sendInfo("You have no active device to play to...")
							continue
						}

						endTime := time.Now()
						currentHostTime := hostState.Progress + int(endTime.Sub(startTime).Milliseconds())

						// Determines whether the track IDs match and whether track progress match
						// If the client is not playing then no need to id match
						var IDmatch bool = false
						if clientState.Item != nil {
							IDmatch = clientState.Item.ID == hostState.Item.ID
						}
						ProgressMatch := ws.Abs(clientState.Progress-currentHostTime) < 5000 // 5 second tolerance

						Log.Trace().Str("Host", s.host.name).Str("Client", client.name).
							Bool("IDMatch", IDmatch).Bool("Progress Match", ProgressMatch).
							Int("Host Progress", currentHostTime).Int("Client Progress", clientState.Progress)

						// If only the progress does not match then only change the seek position
						if IDmatch && !ProgressMatch {
							err = client.spotifyClient.Seek(currentHostTime)
							if err != nil {
								Log.Warn().Str("Username", client.name).Bool("Host", false).Err(err).Msg("Spotify player seek error")
								continue
							}
						}

						// Otherwise change the track and the progress
						if !IDmatch {
							var opts = &spotify.PlayOptions{
								PositionMs: currentHostTime,
								URIs:       []spotify.URI{hostState.Item.URI},
							}

							err = client.spotifyClient.PlayOpt(opts)
							if err != nil {
								Log.Warn().Str("Username", client.name).Bool("Host", false).Err(err).Msg("Spotify player handleSync error")
								continue
							}

							client.sendInfo("Track changed to: " + hostState.Item.Name)
						}
					}
				}
			}

		case <-s.quit:
			ticker.Stop()
			return
		}
	}
}
