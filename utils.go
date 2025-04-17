package main

import (
	"log/slog"
	"os"
)

func must(logger *slog.Logger, err error, msg string) {
	if err != nil {
		logger.Error(msg, "err", err)
		os.Exit(1)
	}
}

func iterateReverse[T any](slice []T, fn func(T)) {
	for i := len(slice) - 1; i >= 0; i-- {
		fn(slice[i])
	}
}

func dereferenceOrDefault[T any](ptr *T, fallback T) T {
	if ptr == nil {
		return fallback
	}
	return *ptr
}
