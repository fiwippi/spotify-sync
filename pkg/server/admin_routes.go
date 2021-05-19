package server

import (
	ws "github.com/fiwippi/spotify-sync/pkg/shared"
	"github.com/gin-gonic/gin"
)

// Route for creating a new user in the database
func createUser(c *gin.Context) {
	var r ws.Request
	err := c.BindJSON(&r)
	if err != nil {
		// If the JSON cannot be unmarshaled then bad request
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Request incorrect"})
		return
	}

	// Bad request if no username or password for new user
	if len(r.NewName) == 0 || len(r.NewPassword) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include user and password"})
		return
	}

	// Ensures keys are valid
	if !(r.AdminKey == adminKey) {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
		return
	}

	err = dbSaveUser(&entry{Name: r.NewName, Password: ws.HashPassword(r.NewPassword)}, false)
	if err != nil {
		Log.Error().Err(err).Msg("Error creating user")
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error creating user"})
		return
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
		return
	}

	// Bad request if no username
	if len(r.CurrentName) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include username"})
		return
	}

	// Ensures keys are valid
	if r.AdminKey != adminKey {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
		return
	}

	err = dbDeleteUser(&entry{Name: r.CurrentName})
	if err != nil {
		Log.Error().Err(err).Msg("Error deleting user")
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error deleting user"})
		return
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
		return
	}

	// Bad request if no data
	if len(r.CurrentName) == 0 {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include username"})
		return
	}

	// Ensures keys are valid
	if r.AdminKey != adminKey {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
		return
	}

	err = dbUpdateUser(&entry{Name: r.CurrentName, NewName: r.NewName, Password: ws.HashPassword(r.NewPassword)})
	if err != nil {
		Log.Error().Err(err).Msg("Error updating user")
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error deleting user"})
		return
	}

	Log.Debug().Str("Old Username", r.CurrentName).Str("Username", r.NewName).Msg("Updated username data")
	c.JSON(200, ws.Response{Success: true, Error: ""})
}

// Route for creating a new user in the database
func viewDB(c *gin.Context) {
	var r ws.Request
	err := c.BindJSON(&r)
	if err != nil {
		// If the JSON cannot be unmarshaled then bad request
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Request incorrect"})
		return
	}

	// Ensures keys are valid
	if !(r.AdminKey == adminKey) {
		c.AbortWithStatusJSON(400, ws.Response{Success: false, Error: "Must include valid admin key"})
		return
	}

	s, err := dbViewAll()
	if err != nil {
		Log.Error().Err(err).Msg("Error viewing db")
		c.AbortWithStatusJSON(500, ws.Response{Success: false, Error: "Error viewing db"})
		return
	}

	c.JSON(200, gin.H{"success": true, "error": "", "db": s})
}
