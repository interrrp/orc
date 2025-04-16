package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

type ServiceManager struct {
	logger   *slog.Logger
	services []*service
}

type service struct {
	logger  *slog.Logger
	config  ServiceConfig
	process *exec.Cmd
}

func NewServiceManager(logger *slog.Logger, configs []ServiceConfig) *ServiceManager {
	var services []*service
	for _, config := range configs {
		services = append(services, &service{logger: logger, config: config})
	}
	return &ServiceManager{logger, services}
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

func (s *service) start() error {
	s.logger.Info("starting service", "name", s.config.Name, "command", s.config.Command)

	process := exec.Command(s.config.Command)
	s.process = process

	if s.config.LogFile != "" {
		logFile, err := os.Create(s.config.LogFile)
		if err != nil {
			return fmt.Errorf("failed to open log file %s for service %s: %w",
				s.config.LogFile, s.config.Name, err)
		}
		process.Stdout = logFile
		process.Stderr = logFile
	}

	if err := process.Start(); err != nil {
		return fmt.Errorf("failed to start service %s: %w", s.config.Name, err)
	}

	return nil
}

func (s *service) stop() error {
	if s.process == nil || s.process.Process == nil {
		s.logger.Warn("attempted to stop inactive service")
		return nil
	}

	s.logger.Info("stopping service", "name", s.config.Name)
	if err := s.process.Process.Signal(syscall.SIGTERM); err != nil {
		s.logger.Error("failed to send SIGTERM", "service", s.config.Name, "err", err)
		return err
	}

	if err := s.process.Wait(); err != nil {
		s.logger.Error("failed to wait for process to stop", "service", s.config.Name, "err", err)
		return err
	}

	return nil
}
