package server

import (
	"github.com/gorilla/websocket"
)

// Upgrades the http connection to a websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}