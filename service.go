package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

type service struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
	LogFile string `toml:"log_file"`
	process *exec.Cmd
}

func startServices(services []service) error {
	for i, svc := range services {
		slog.Info("starting service", "name", svc.Name, "command", svc.Command)

		process := exec.Command(svc.Command)
		services[i].process = process

		if svc.LogFile != "" {
			logFile, err := os.Create(svc.LogFile)
			if err != nil {
				return fmt.Errorf("failed to open log file %s for service %s: %w", svc.LogFile, svc.Name, err)
			}
			process.Stdout = logFile
			process.Stderr = logFile
		}

		if err := process.Start(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.Name, err)
		}
	}
	return nil
}

func stopServices(services []service) error {
	var lastErr error
	for i := len(services) - 1; i >= 0; i-- {
		svc := services[i]
		if svc.process == nil || svc.process.Process == nil {
			continue
		}

		slog.Info("stopping service", "name", svc.Name)
		if err := svc.process.Process.Signal(syscall.SIGTERM); err != nil {
			slog.Error("failed to send SIGTERM", "service", svc.Name, "err", err)
			lastErr = err
			continue
		}

		if err := svc.process.Wait(); err != nil {
			slog.Error("failed to wait for process to stop", "service", svc.Name, "err", err)
			lastErr = err
		}
	}
	return lastErr
}
