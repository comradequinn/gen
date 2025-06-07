// Package log contains log utilities
package log

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	logger *slog.Logger
	w      func(string, ...any)
)

// Init configures a logger for use
func Init(debug bool, writer func(string, ...any)) {
	if logger != nil {
		panic("log can only be initialised once")
	}

	w = writer
	logLevel := slog.LevelInfo

	if debug {
		logLevel = slog.LevelDebug
	}

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
}

// FatalfIf logs an error, prints it to stdout and exits with code
func FatalfIf(condition bool, format string, v ...any) {
	if !condition {
		return
	}
	logger.Error("process terminated due to error", "err", fmt.Errorf(format, v...))
	w(format+"\n", v...)
	os.Exit(1)
}

func DebugPrintf(msg string, args ...any) {
	logger.Debug(msg, args...)
}
