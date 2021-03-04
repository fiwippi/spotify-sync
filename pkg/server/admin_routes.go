package server

import (
	"github.com/gin-gonic/gin"
	"log"
	ws "spotify-sync/pkg/shared"
)

// These are all router routes for providing admin functionality

// Route for creating a new user in the database
func createUser(c *gin.Context) {
	var r ws.Request
	err := c.BindJSON(&r)
	if err != nil {
		// If the JSON cannot be unmarshaled then bad request
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Request incorrect"})
	}

	// Bad request if no username or password for new user
	if len(r.NewName) == 0 || len(r.NewPassword) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include user and password"})
	}

	// Ensures keys are valid
	if !(r.ServerKey == serverKey || r.AdminKey == adminKey) {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid server or admin key"})
	}

	err = dbSaveUser(&entry{Name: r.NewName, Password: ws.HashPassword(r.NewPassword)}, false)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error creating user"})
	}

	Log.Debug().Str("Username", r.NewName).Msg("Created user")
	c.JSON(200, ws.Response{Success: true, Error: ""})
}

// Route for deleting a user in the database
func deleteUser(c *gin.Context) {
	var r ws.Request
	err := c.BindJSON(&r)
	if err != nil {
		// If the JSON cannot be unmarshaled then bad request
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Request incorrect"})
	}

	// Bad request if no username
	if len(r.CurrentName) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include username"})
	}

	// Ensures keys are valid
	if r.AdminKey != adminKey {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
	}

	err = dbDeleteUser(&entry{Name: r.CurrentName})
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error deleting user"})
	}

	Log.Debug().Str("Username", r.NewName).Msg("Deleted user")
	c.JSON(200, ws.Response{Success: true, Error: ""})
}

// Route for updating a user's data in the database
func updateUser(c *gin.Context) {
	var r ws.Request
	err := c.BindJSON(&r)
	if err != nil {
		// If the JSON cannot be unmarshaled then bad request
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Request incorrect"})
	}

	// Bad request if no data
	if len(r.CurrentName) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include username"})
	}

	// Ensures keys are valid
	if r.AdminKey != adminKey {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
	}

	err = dbUpdateUser(&entry{Name: r.CurrentName, NewName: r.NewName, Password: ws.HashPassword(r.NewPassword)})
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error deleting user"})
	}

	Log.Debug().Str("Old Username", r.CurrentName).Str("Username", r.NewName).Msg("Updated username data")
	c.JSON(200, ws.Response{Success: true, Error: ""})
}
