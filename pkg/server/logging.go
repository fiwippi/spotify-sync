package server

import (
	"github.com/rs/zerolog"
	"os"
)

var Log zerolog.Logger

// Creates the server logger
func createLogger(logLevel string) (zerolog.Logger, error) {
	// Set the log level
	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create the data dir
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", os.ModeDir)
	}

	// Creates a file for writing logs to
	logFile, err := os.OpenFile("data/server.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return zerolog.Logger{}, err
	}

	// Creates the console logger
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}

	// Combine separate loggers into the main logger
	w := zerolog.MultiLevelWriter(consoleWriter, logFile)
	mainLog := zerolog.New(w).With().Timestamp().Logger()

	return mainLog, nil
}
