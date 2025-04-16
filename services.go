package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

func startServices() {
	for i, service := range cfg.Services {
		slog.Info("starting service", "name", service.Name, "command", service.Command)

		process := exec.Command(service.Command)
		cfg.Services[i].process = process

		if service.LogFile != "" {
			logFile, err := os.Create(service.LogFile)
			must(err, "failed to open service log file")
			process.Stdout = logFile
			process.Stderr = logFile
		}

		_ = process.Start()
	}
}

func stopServices() {
	for i := len(cfg.Services) - 1; i >= 0; i-- {
		service := cfg.Services[i]
		slog.Info("stopping service", "name", service.Name)
		must(service.process.Process.Signal(syscall.SIGTERM), "failed to send SIGTERM to service")
		must(service.process.Wait(), "failed to wait for process to stop")
	}
}
