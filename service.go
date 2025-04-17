package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type ServiceManager struct {
	logger   *slog.Logger
	services []*service
}

func NewServiceManager(logger *slog.Logger, serviceDir string) (*ServiceManager, error) {
	var services []*service

	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("reading service directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			services = append(services, &service{
				logger:     logger,
				name:       entry.Name(),
				scriptPath: filepath.Join(serviceDir, entry.Name()),
			})
		}
	}

	return &ServiceManager{logger, services}, nil
}

func (sm *ServiceManager) StartAll() error {
	var lastErr error
	for _, svc := range sm.services {
		if err := svc.start(); err != nil {
			sm.logger.Error("failed to start service", "err", err)
			lastErr = err
		}
	}
	return lastErr
}

func (sm *ServiceManager) StopAll() error {
	var lastErr error
	iterateReverse(sm.services, func(svc *service) {
		if err := svc.stop(); err != nil {
			sm.logger.Error("failed to stop service", "err", err)
			lastErr = err
		}
	})
	return lastErr
}

type service struct {
	logger     *slog.Logger
	name       string
	scriptPath string
	process    *exec.Cmd
}

func (svc *service) start() error {
	svc.logger.Info("starting service", "name", svc.name)

	cmd := exec.Command("/bin/sh", "-c", ". "+svc.scriptPath+" && start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting service: %w", err)
	}
	svc.process = cmd

	return nil
}

func (svc *service) stop() error {
	if svc.process == nil || svc.process.Process == nil {
		return nil
	}

	svc.logger.Info("stopping service", "name", svc.name)

	stopCmd := exec.Command("/bin/sh", "-c", ". "+svc.scriptPath+" && stop")
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr

	if err := stopCmd.Run(); err != nil {
		svc.logger.Warn("failed to run stop function", "err", err)
	}

	if err := svc.process.Process.Signal(syscall.SIGTERM); err != nil {
		svc.logger.Warn("failed to send SIGTERM, killing process", "err", err)
		if err := svc.process.Process.Kill(); err != nil {
			svc.logger.Warn("failed to kill process", "err", err)
		}
	}

	if err := svc.process.Wait(); err != nil {
		svc.logger.Warn("failed to wait for process to exit", "err", err)
	}

	svc.process = nil

	return nil
}
