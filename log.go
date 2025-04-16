package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func setUpLogging() {
	opts := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
	}
	handler := tint.NewHandler(os.Stderr, opts)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	clearScreen()
}

func clearScreen() {
	fmt.Print("\033c")
}
