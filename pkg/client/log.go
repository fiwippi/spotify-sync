package client

import (
	"log"
	"os"
)

var Log *log.Logger = CreateLogger()

// Creates the logger which writes to a log file
func CreateLogger() *log.Logger {
	f, err := os.OpenFile("client.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
}