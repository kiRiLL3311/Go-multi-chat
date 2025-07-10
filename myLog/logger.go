package myLog

import (
	"log/slog"
	"os"
)

var logFile *os.File

type LogLevel struct {
	INFO    string
	WARNING string
	ERROR   string
}

func InitLog() {
	LOG_FILE := os.Getenv("LOG_FILE")

	var err error
	logFile, err = os.OpenFile(LOG_FILE, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	handlerOptions := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewTextHandler(logFile, handlerOptions))
	slog.SetDefault(logger)
}

func MidLog(username string, msg string, logType string) {
	switch logType {
	case "I":
		slog.With("username", username).Info(msg)
	case "W":
		slog.With("username", username).Warn(msg)
	case "E":
		slog.With("username", username).Error(msg)
	default:
		slog.With("username", username).Error("Undefined error, no such log level")

	}
}

func CloseLog() {
	if logFile != nil {
		logFile.Close()
	}
}
