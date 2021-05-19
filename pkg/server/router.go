package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Admin key allows deeper access to server functions, i.e. deleting accounts, updating account data
var adminKey string

// Middleware to ensure an authenticator has been generated
func authGenerated() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authCreated {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"msg": "Cannot process requests currently"})
		}
	}
}

// Creates the server object
func setupServer(aK, id, secret, redirect, port, logLevel string) (*http.Server, error) {
	var err error

	// Create the logger
	Log, err = createLogger(logLevel)
	if err != nil {
		return nil, err
	}

	// Set the admin key
	adminKey = aK

	// connect to the database
	db, err = bolt.Open("data/spotify.db", 0666, nil)
	if err != nil {
		return nil, err
	}

	// Guarantees main user bucket exists
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Generate the router
	router := gin.Default() // Default router  includes logging and recovery middleware
	router.Use(authGenerated())

	// Setup templating
	tmpl, err := template.ParseFS(efs, "static/admin.tmpl")
	if err != nil {
		return nil, err
	}
	router.SetHTMLTemplate(tmpl)

	// Add routes
	router.GET("/admin", admin)
	router.GET("/favicon.ico", favicon)
	router.GET("/spotify-callback", spotifyCallback)
	router.GET("/shared", processIncomingUserWebsocket)
	router.POST("/api/create-user", createUser)
	router.POST("/api/delete-user", deleteUser)
	router.POST("/api/update-user", updateUser)
	router.POST("/api/view-db", viewDB)

	// Generate the spotify auth object
	err = generateAuth(id, secret, redirect)
	if err != nil {
		return nil, err
	}

	// Attach the router to a http server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	srv.RegisterOnShutdown(func() {
		// Close each session
		for s := range sessions {
			if sessions[s] != nil {
				sessions[s].close()
			}
		}

		// Disconnect all users
		for u := range connectedUsers {
			u.disconnect()
		}
	})

	return srv, nil
}

// Run a server
func Run(aK, id, secret, redirect, port, mode, logLevel string, refresh time.Duration) error {
	// Set the refresh
	syncRefresh = refresh

	if mode != "release" && mode != "debug" {
		return errors.New("invalid mode, must be \"release\" or \"debug\"")
	}
	gin.SetMode(mode)

	// Create the server
	srv, err := setupServer(aK, id, secret, redirect, port, logLevel)
	if err != nil {
		return err
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	Log.Warn().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(ctx); err != nil {
		Log.Fatal().Err(err).Msg("Server Shutdown")
	}
	Log.Warn().Msg("Server exiting")

	return nil
}
