package client

import (
	"errors"
	"fmt"
	ws "github.com/fiwippi/spotify-sync/pkg/shared"
	"strings"
)

//// READING MESSAGES

// Reads messages from the connection and processes them
func (c *Client) readPump() {
	var msg ws.Message
	for {
		// Retrieve the ws.Message struct from the connection
		err := c.conn.ReadJSON(&msg)
		Log.Printf("Incoming Message: %+v\n", msg)
		if err != nil {
			// Error here means the connection is closed
			Log.Println(fmt.Sprintf("read: %s", err.Error()))

			// Causes disconnect handling to occur
			c.notifyDone()

			return
		}

		// Process the struct and call the appropriate command
		err = c.processMsg(msg)
		if err != nil {
			Log.Printf("processing: %s", err.Error())
			return
		}

		// Redraws the gui to show changes
		gCtx.app.Draw()
	}
}

// Processes messages and calls the relevant function
func (c *Client) processMsg(m ws.Message) error {
	var err error

	if !ws.OPCODES.Has(m.Op) {
		return errors.New("Opcode doesn't exist")
	}

	switch cmd := m.Op; cmd {
	case "AUTH":
		err = c.cmdAuth(&m)
	case "INFO":
		err = c.cmdInfo(&m)
	case "LOGIN":
		err = c.cmdLogin()
	case "USERS":
		err = c.cmdUsers(&m)
	case "MSG":
		err = c.cmdMsg(&m)
	default:
		Log.Printf("Could not process msg: %+v\n", m)
	}

	if err != nil {
		return err
	}
	return nil
}

//// WRITING MESSAGES

// Takes the user input and converts it into the ws.Message format
// if applicable and then writes it over the connection to the server
func (c *Client) writeMsg(text string) {
	// Clean up the text
	text = strings.TrimSuffix(text, "\n")
	if !strings.Contains(text, ",") { // We can use Contains because commas are only used as delimiters
		text += ","
	}

	// Message must contain at least 2 sections, the second one can be blank in cases of the CREATE opcode, etc.
	sections := strings.Split(text, ",")
	if len(sections) < 2 {
		// Do not attempt to create message if not enough sections
		Log.Printf("write: Input ignored '%s'\n", text)
		return
	}

	// Generate the message and send it
	msg, err := c.buildMsg(sections)
	if err != nil {
		Log.Printf("write: %s", err.Error())
		return
	}

	// Sends the message over the websocket connection
	err = c.conn.WriteJSON(msg)
	if err != nil {
		Log.Printf("write: %s", err.Error())
		return
	}
}

// Converts the input text from the end user into the message struct
// for sending to the server. Also processes client-side opcodes, i.e. exit/quit
func (c *Client) buildMsg(sections []string) (ws.Message, error) {
	// Ensures the opcode is valid
	op := strings.ToUpper(sections[0])
	if !ws.OPCODES.Has(op) {
		return ws.Message{}, errors.New("Opcode doesn't exist")
	}

	// Builds the message
	msg := ws.Message{
		Op:        op,
		Args:      nil,
		Body:      strings.Join(sections[1:], ","),
		Timestamp: ws.CurrentTime(),
	}

	// Processes client side opcodes
	switch op {
	case "EXIT":
		Log.Println("Sending struct to done channel due to EXIT opcode")
		c.notifyDone()
	case "QUIT":
		Log.Println("Sending struct to done channel due to QUIT opcode")
		c.notifyDone()
	}
	return msg, nil
}
