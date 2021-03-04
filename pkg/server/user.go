package server

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
	sets "spotify-sync/pkg/set"
	ws "spotify-sync/pkg/shared"
	"strings"
	"sync"
	"time"
)

// Set to hold the connected usernames, (not the user object)
var ids = sets.NewSet()
// Keeps track of all the connected users (the user objects)
var connectedUsers = make(map[*user]bool)
// Channel used to send the spotify.Client created in the SpotifyCallback route
// which is then sent to the user handshake where it is assigned to the user
var clientChans = make(map[string]chan spotify.Client, 1)
// Channel used to send the oauth2.Token created in the SpotifyCallback route
// which is then sent to the user handshake where it is assigned to the user
var tokenChans = make(map[string]chan *oauth2.Token, 1)

// Active users connected to the server
type user struct {
	mutex         sync.Mutex           // Locks writing to websocket conn
	name          string               // Identifies the user in the database
	r             *http.Request        // Request used to upgrade the user connection
	w             http.ResponseWriter  // Response writer used to upgrade the user connection
	conn          *websocket.Conn      // Servers shared connection to the user client
	spotifyClient spotify.Client       // The spotify client used to control the user's spotify
	spotifyData   *spotify.PrivateUser // Holds data about the user
	s             *session             // The current session the user is connected to
	token         *oauth2.Token        // Token used to refresh access to the client
}

// Upgrades user to shared connection (websocket) from a http connection
func (u *user) upgrade() error {
	if u.r == nil {
		return errors.New("user has no http.request")
	} else if u.w == nil {
		return errors.New("user has no http.responsewriter")
	}

	conn, err := upgrader.Upgrade(u.w, u.r, nil)
	if err != nil {
		return err
	}

	u.conn = conn
	return nil
}

// Disconnects a user from the server
func (u *user) disconnect() {
	Log.Info().Str("Username", u.name).Msg("Disconnecting user")

	// Stop keeping track of the user
	ids.Remove(u.name)
	delete(connectedUsers, u)

	// Close relevant channels
	clientChan := clientChans[u.name]
	if clientChan != nil && !isSpotifyClientChanClosed(clientChan) {
		close(clientChan)
	}
	tokenChan := tokenChans[u.name]
	if tokenChan != nil && !isSpotifyTokenChanClosed(tokenChan) {
		close(tokenChan)
	}

	// Determine if user is connected to session
	if u.s != nil {
		// If the user is hosting a session then close the session, otherwise remove them from the members
		if u.s.host == u {
			u.s.close()
			_, ok := sessions[u.name];
			if ok {
				delete(sessions, u.name);
			}
		} else {
			//log.Printf("%+v unregistering %+v", u.s, u)
			// If the session's unregister channel is open then unregister the user
			if !unregisterChannelClosed(u.s.unregister) {
				u.s.unregister <- u
			}
			u.s = nil
			//log.Printf("%+v unregistering %+v", u.s, u)
		}
	}

	// If the user still has an active connection then disconnect them
	if u.conn != nil {
		// Inform user they are being disconnected
		_ = u.sendInfo("Disconnection occurring")

		// Close the shared connection
		u.conn.Close()
	}
}

// Performs the handshake procedure
func (u *user) handshake() error {
	// Ask the user to send login credentials
	msg :=  &ws.Message{
		Op:   "LOGIN",
		Args: nil,
		Body: "",
		Timestamp: ws.CurrentTime(),
	}
	err := u.WriteJSON(msg)
	if err != nil {
		return err
	}

	// Read the username and password
	var username, password string
	errChan := make(chan error)
	go func() {
		err = u.conn.ReadJSON(msg)
		errChan <- err
	}()
	select {
	case err = <- errChan:
		reply := strings.Split(msg.Body, ",")
		if len(reply) != 2 {
			return errors.New("No username or password")
		}
		username, password = reply[0], reply[1]
	case <- time.After(1 * time.Minute):
		close(errChan)
		return errors.New("Timeout for authorising access to account (client)")
	}

	// Verify the credentials exist
	e, err := dbViewUser(username)
	if err != nil {
		return errors.New("User not retrieved successfully from database, " + err.Error())
	}
	if e.Password != ws.HashPassword(password) {
		return errors.New("Password incorrect")
	}

	// Disconnect if user is already connected
	u.name = username
	if ids.Has(u.name) {
		return errors.New("User already connected")
	}
	ids.Add(u.name)
	connectedUsers[u] = true

	// Try and recreate client
	if e.Token != "" {
		var tkn *oauth2.Token
		err := json.Unmarshal([]byte(e.Token), &tkn)
		if err != nil {
			return errors.New("Cannot parse token from db: " + err.Error())
		}

		u.token = tkn
		u.spotifyClient = auth.NewClient(u.token)
	} else {
		clientChans[u.name] = make(chan spotify.Client, 1)
		tokenChans[u.name] = make(chan *oauth2.Token, 1)

		// Tell user to authenticate via auth URL sent to them
		msg =  &ws.Message{
			Op:   "AUTH",
			Args: nil,
			Body: auth.AuthURL(u.name),
			Timestamp: ws.CurrentTime(),
		}
		err = u.WriteJSON(msg)
		if err != nil {
			return err
		}

		// First we receive the token from the spotify callback function
		select {
		case t := <- tokenChans[u.name]:
			u.token = t
		case <- time.After(5 * time.Minute):
			return errors.New("Timeout for authorising access to account (client)")
		}

		// Second we receive the spotify client from the spotify callback function
		select {
		case sc := <-clientChans[u.name]:
			u.spotifyClient = sc
		case <- time.After(5 * time.Second):
			return errors.New("Timeout for authorising access to account (token)")
		}

		// Save the created token to the db so the spotify client can be recreated for the user on connection
		tokenBytes, err := json.Marshal(u.token)
		if err != nil {
			return errors.New("Cannot encode json token into byte string: " + err.Error())
		}

		e = &entry{
			Name:     u.name,
			Password: ws.HashPassword(password),
			Token:    string(tokenBytes),
		}

		err = dbSaveUser(e, true)
		if err != nil {
			return errors.New("Failed saving token to db: " + err.Error())
		}
	}

	// Inform user of successful handshake
	err = u.sendInfo("Spotify client authorised, handshake successful!")
	if err != nil {
		return err
	}
	return nil
}

// Reads messages from the connection and processes them
func (u *user) readPump() {
	var msg ws.Message
	for {
		// Retrieve the ws.Message struct from the connection
		err := u.conn.ReadJSON(&msg)
		if err != nil {
			Log.Debug().Err(err).Str("Username", u.name).Msg("Read error")
			u.disconnect()
			return
		}

		// Process the struct and call the appropriate command
		err = u.processMsg(msg)
		if err != nil {
			Log.Debug().Err(err).Str("Username", u.name).Msg("Processing error")
			return
		}
	}
}

// Tells the user client to refresh all users
func (u *user) clearUserList() {
	msg :=  &ws.Message{
		Op:   "USERS",
		Args: nil,
		Body: "",
		Timestamp: ws.CurrentTime(),
	}

	_ = u.WriteJSON(msg)
}

// Processes messages and calls the relevant function
func (u *user) processMsg(m ws.Message) error {
	var err error

	if !ws.OPCODES.Has(m.Op) {
		return errors.New("Opcode doesn't exist")
	}

	Log.Info().Str("OPCODE", m.Op).Str("Username", u.name).Msg(m.Body)

	switch cmd := m.Op; cmd {
	case "CREATE":
		err = u.cmdCreate(&m)
	case "ID":
		err = u.cmdID(&m)
	case "JOIN":
		err = u.cmdJoin(&m)
	case "DISCONNECT":
		err = u.cmdDisconnect(&m)
	case "MSG":
		err = u.cmdMsg(&m)
	case "HELP":
		err = u.cmdHelp(&m)
	default:
		Log.Warn().Str("OPCODE", m.Op).Msg("Could not process message")
	}

	if err != nil {
		return err
	}
	return nil
}

// Sends a message and avoids concurrent writes
func (u *user) WriteJSON(m *ws.Message) error {
	if u.conn != nil {
		u.mutex.Lock()
		err := u.conn.WriteJSON(m)
		u.mutex.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

// Sends an INFO message to the user (message from server)
func (u *user) sendInfo(text string) error {
	msg :=  &ws.Message{
		Op:   "INFO",
		Args: nil,
		Body: text,
		Timestamp: ws.CurrentTime(),
	}
	
	err := u.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}

// Sends a MSG message to the user (message from other clients)
func (u *user) sendMsg(text string) error {
	msg :=  &ws.Message{
		Op:   "MSG",
		Args: []string{u.name},
		Body: text,
		Timestamp: ws.CurrentTime(),
	}

	err := u.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}