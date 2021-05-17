package server

import (
	"errors"
	ws "github.com/fiwippi/spotify-sync/pkg/shared"
	"strings"
)

var helpMsg = `
###### HELP ######
CREATE = Create a session
JOIN = Join a session someone has created e.g. "join,username"
EXIT/QUIT = Disconnect from the server
DISCONNECT = Leave the session
ID = Displays the ID of the current session
MSG = Send a message to other users in the same session e.g. "msg,change the song?""`

// Sends a help message to the user
func (u *user) cmdHelp(m *ws.Message) error {
	return u.sendInfo(helpMsg)
}

// Sends a message to all clients in the session
func (u *user) cmdMsg(m *ws.Message) error {
	if u.s != nil {
		u.s.broadcast <- m.Body
	} else {
		_ = u.sendInfo("No session to send message to")
	}

	return nil
}

// Sends the session ID to the user if they're in one
func (u *user) cmdID(m *ws.Message) error {
	if u.s == nil {
		return u.sendInfo("ID: N/A")
	}
	return u.sendInfo("ID: " + u.s.host.name)
}

// Creates a new session
func (u *user) cmdCreate(m *ws.Message) error {
	// Ensures user is not already in a session
	if u.s != nil {
		u.sendInfo("Cannot create a session while you're already in one")
		return nil
	}

	// Create the session
	sessions[u.name] = newSession(u)

	// Notify of success
	err := u.sendInfo("Session created for: " + u.name)
	if err != nil {
		return err
	}

	// Send the user list
	err = sessions[u.name].sendUserUpdate()
	if err != nil {
		return err
	}

	Log.Info().Str("Username", u.name).Msg("Session created")

	// Start the session
	go sessions[u.name].handleChannels()
	go sessions[u.name].handleSync()

	return nil
}

// Adds a user to the session
func (u *user) cmdJoin(m *ws.Message) error {
	// Ensures user is not already in a session
	if u.s != nil {
		u.sendInfo("Cannot join a session while you're already in one")
		return nil
	}

	// Get the id of the session to join
	idArray := strings.Split(m.Body, ",")
	if len(idArray) == 0 {
		return errors.New("Bad message content for joining session")
	}
	id := idArray[0]

	// Check if the session exists
	var text string
	if _, ok := sessions[id]; ok {
		text = "Session (" + id + ") joined by: " + u.name
		sessions[id].register <- u
		u.s = sessions[id]
	} else {
		text = "Cannot join session (" + id + ") for: " + u.name
	}

	Log.Info().Str("Username", u.name).Msg(text)

	// Notify of success or fail
	err := u.sendInfo(text)
	if err != nil {
		return err
	}

	return nil
}

// Disconnects a user from the session
func (u *user) cmdDisconnect(m *ws.Message) error {
	// Check if the session exists
	var text string
	if u.s != nil {
		isHost := u.s.host == u
		sessionName := u.s.host.name
		if isHost {
			u.s.close()
			u.s = nil
			_, ok := sessions[u.name]
			if ok {
				delete(sessions, u.name)
			}
		} else {
			u.s.unregister <- u
			u.s = nil
		}
		text = "Session (" + sessionName + ") left for: " + u.name
		u.clearUserList()
		Log.Info().Str("Username", u.name).Bool("Is Host", isHost).Msg("Disconnected from session")
	} else {
		text = "Not in a session"
	}

	// Notify of success or fail
	err := u.sendInfo(text)
	if err != nil {
		return err
	}

	return nil
}
