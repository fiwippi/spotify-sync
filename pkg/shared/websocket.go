package ws

import (
	sets "spotify-sync/pkg/set"
)

// Set of all accepted opcodes
var OPCODES *sets.Set = generateOpcodesSet()

// The message sent over the websocket connection to the server
type Message struct {
	Op        string   `json:"op"`        // Name of the command
	Args      []string `json:"args"`      // Extra Args for the command, supplied if needed e.g. MSG opcode
	Body      string   `json:"body"`      // Body of the command
	Timestamp string   `json:"timestamp"` // Timestamp of the message
}

// Function to create all the opcodes
func generateOpcodesSet() *sets.Set {
	op := sets.NewSet()

	// Opcodes used by the server/client internally
	op.Add("AUTH", "INFO", "LOGIN", "USERS")
	// End-user opcodes
	op.Add("CREATE", "JOIN", "DISCONNECT", "ID", "MSG", "HELP", "EXIT", "QUIT")

	return op
}