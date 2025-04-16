package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	logger := CreateLogger()
	system, err := NewSystem(logger)
	must(logger, err, "failed to initialize")
	must(logger, system.Start(), "failed to start")
	must(logger, system.Stop(), "failed to stop")
}

func CreateLogger() *slog.Logger {
	opts := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
	}
	handler := tint.NewHandler(os.Stderr, opts)
	return slog.New(handler)
}
