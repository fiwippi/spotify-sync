package client

import (
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"net/url"
	"os"
	"os/signal"
	"time"
)

// The gui app object (this is the root object of the gui)
var app *tview.Application

// Client used to connect to the spotify sync server
type Client struct {
	done      chan struct{}   // Channel for notifying the client is done reading messages from the shared conn
	url       url.URL         // URL the client will connect to via HTTP and then upgrade to websocket
	conn      *websocket.Conn // Websocket connection used to connect to the server
	interrupt chan os.Signal  // Channel to signal the client to close the socket connection
}

// Create the client object with its respective channels
func NewClient() *Client {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	return &Client{
		done:      make(chan struct{}),
		interrupt: interrupt,
	}
}

// Builds the gui and then runs it, returns an error on failure
func (c *Client) Run() error {
	app = c.createGUI()
	if err := app.Run(); err != nil {
		return err
	}
	return nil
}

// Changes the address the client will attempt to connect to via websocket
func (c *Client) changeAddress(addr string) {
	scheme := "wss"
	if !useSSL {
		scheme = "ws"
	}

	c.url = url.URL{Scheme: scheme, Host: addr, Path: "/shared"}
}

// Dials the shared connection and connects to the sync server
func (c *Client) connect() error {
	var err error
	Log.Println("Dialing to:", c.url.String())
	c.conn, _, err = websocket.DefaultDialer.Dial(c.url.String(), nil)
	if err != nil {
		return err
	}
	return nil
}

// Close the client's websocket connection to the server
func (c *Client) disconnect() error {
	Log.Println("Closing websocket connection")
	err := c.conn.Close()
	if err != nil {
		return err
	}
	Log.Println("Websocket connection closed")
	return nil
}

// Sends an empty struct to the done channel to cause a shutdown
func (c *Client) notifyDone() {
	select {
	case c.done <- struct{}{}:
	default:
		Log.Println("Done channel full. Discarding value")
	}
}

// Wait for client to stop waiting for read messages (signified by
// closing of done channel) or due to an external interrupt i.e. Ctrl+C
func (c *Client) handleShutdown() {
	defer func() {
		c.done = make(chan struct{})
		err := c.disconnect()
		if err != nil {
			Log.Println("Error closing websocket conn: " + err.Error())
		}
	}()

	select {
	// Manual shutdown by user or through (unexpected) closed websocket connection
	case <-c.done:
		Log.Println("Manual shutdown / Websocket err")
		// Clean up the old text boxes
		gCtx.users.Clear().SetText("USERS")
		gCtx.chatlog.Clear()

		// Go back to home screen if not shutting down
		gCtx.pages.SwitchToPage("disconnected")

		return
	// Shutdown through interrupt
	case <-c.interrupt:
		Log.Println("Shutdown through interrupt")
		// Cleanly close the connection by sending a close message and then waiting (with timeout) for the server to close the connection.
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			Log.Printf("write close: %s", err.Error())
			return
		}
		select {
		case <-c.done:
		case <-time.After(time.Second):
		}
		return
	}
}
