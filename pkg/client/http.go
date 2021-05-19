package client

import (
	"bytes"
	"encoding/json"
	"errors"
	ws "github.com/fiwippi/spotify-sync/pkg/shared"
	"net/http"
	"net/url"
)

// Sends a http request to the server
func (c *Client) sendRequest(r *ws.Request, endpoint string) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	resp, err := http.Post((&url.URL{Scheme: scheme, Host: c.url.Host}).String()+endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	Log.Println("Response received")

	var response = new(ws.Response)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}
	Log.Printf("Response decoded: %+v\n", response)

	if !response.Success {
		return errors.New(response.Error)
	}
	return nil
}

// Request wrapper to create a user
func (c *Client) createUser(name, pass, sK, aK string) error {
	return c.sendRequest(&ws.Request{NewName: name, NewPassword: pass, ServerKey: sK, AdminKey: aK}, "/create-user")
}

// Request wrapper to delete a user
func (c *Client) deleteUser(name, aK string) error {
	return c.sendRequest(&ws.Request{CurrentName: name, AdminKey: aK}, "/delete-user")
}

// Request wrapper to update a user's details, i.e. change their name or password
func (c *Client) updateUser(name, newname, pass, aK string) error {
	return c.sendRequest(&ws.Request{CurrentName: name, NewName: newname, NewPassword: pass, AdminKey: aK}, "/update-user")
}
