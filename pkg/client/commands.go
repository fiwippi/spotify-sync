package client

import (
	"fmt"
	"github.com/atotto/clipboard"
	ws "spotify-sync/pkg/shared"
	"strings"
)

//// SERVER to CLIENT opcodes

// Processes the AUTH opcode
func (c *Client) cmdAuth(m *ws.Message) error {
	// Copies the spotify oauth2 url to the clipboard
	err := clipboard.WriteAll(m.Body)
	if err != nil {
		return err
	}

	// Writes the auth url to the chatlog
	text := fmt.Sprintf("The authentication URL has been copied to the clipboard. Please authenticate the client through: %s\n", m.Body)
	gCtx.chatlog.Write([]byte(fmt.Sprintf("[red]%s <SERVER> %s", m.Timestamp, text)))

	return nil
}

// Process the INFO opcode
func (c *Client) cmdInfo(m *ws.Message) error {
	// Writes the info text to the chatlog
	text := fmt.Sprintf("INFO: %s\n", m.Body)
	gCtx.chatlog.Write([]byte(fmt.Sprintf("[red]%s <SERVER> %s", m.Timestamp, text)))

	return nil
}

// Process the MSG opcode
func (c *Client) cmdMsg(m *ws.Message) error {
	// If username cannot be retrieved then write it in red
	var name string = "[red]name error[teal]"
	if m.Args != nil && len(m.Args) > 0 {
		name = m.Args[0]
	}

	// Write the user message to the chatlog
	gCtx.chatlog.Write([]byte(fmt.Sprintf("[teal]%s <%s>: %s", m.Timestamp, name, m.Body)))
	return nil
}

// Processes the USERS opcode
func (c *Client) cmdUsers(m *ws.Message) error {
	// Clears the user box and rewrites the current users to it
	gCtx.users.Clear()
	usersString := "USERS\n\n" + strings.ReplaceAll(m.Body, ",", "\n")
	_, err := gCtx.users.Write([]byte(usersString))
	if err != nil {
		return err
	}

	return nil
}

// Processes the LOGIN opcode, this means the server is asking for the user's login details
func (c *Client) cmdLogin() error {
	c.writeMsg(fmt.Sprintf("login,%s,%s", details.Username, details.Password))

	return nil
}
