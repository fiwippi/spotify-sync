package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server key allows users basic access to server functions, i.e. creating accounts
// Admin key allows deeper access to server functions, i.e. deleting accounts, updating account data
var serverKey, adminKey string

// Middleware to ensure an authenticator has been generated
func authGenerated() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authCreated {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"msg":"Cannot process requests currently"})
		}
	}
}

// Creates the server object
func setupServer(sK, aK, id, secret, redirect, port string) (*http.Server, error) {
	var err error

	// Create the logger
	Log, err = createLogger()
	if err != nil {
		return nil, err
	}

	// Set the server and admin key
	serverKey, adminKey = sK, aK

	// connect to the database
	db, err = bolt.Open("spotify.db", 0666, nil)
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
	router.GET("/spotify-callback", spotifyCallback)
	router.GET("/shared", processIncomingUserWebsocket)
	router.POST("/create-user", createUser)
	router.POST("/delete-user", deleteUser)
	router.POST("/update-user", updateUser)

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
func Run(sK, aK, id, secret, redirect, port string, refresh time.Duration) error {
	// Set the refresh
	syncRefresh = refresh

	// Create the server
	srv, err := setupServer(sK, aK, id, secret, redirect, port)
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