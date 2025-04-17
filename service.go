package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"
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

func (svc *service) start() error {
	svc.logger.Info("starting service", "name", svc.config.Name, "command", svc.config.Command)
	if svc.config.Mode == "oneshot" {
		svc.supervise()
	} else {
		go svc.supervise()
	}
	return nil
}

func (svc *service) supervise() {
	for {
		process := exec.Command(svc.config.Command)
		svc.process = process

		svc.redirectOutputToLogFile()

		if err := process.Start(); err != nil {
			svc.logger.Error("failed to start service", "name", svc.config.Name, "error", err)
			return
		}

		exitCode := svc.waitForExit()

		if svc.shouldStop(exitCode) {
			break
		}
	}
}

func (svc *service) redirectOutputToLogFile() {
	if svc.config.LogFile == "" {
		return
	}

	logFile, err := os.Create(svc.config.LogFile)
	if err != nil {
		svc.logger.Error("failed to open log file", "name", svc.config.Name, "error", err)
		return
	}
	svc.process.Stdout = logFile
	svc.process.Stderr = logFile
}

func (svc *service) waitForExit() int {
	err := svc.process.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return exitCode
}

func (svc *service) shouldStop(exitCode int) bool {
	if exitCode == 0 {
		svc.logger.Info("service exited", "name", svc.config.Name, "exitCode", exitCode)
		return true
	} else {
		svc.logger.Warn("service exited", "name", svc.config.Name, "exitCode", exitCode)

		shouldRestart := dereferenceOrDefault(svc.config.RestartOnFailure, true)
		if shouldRestart {
			svc.logger.Warn("restarting failed service", "name", svc.config.Name)
			time.Sleep(1 * time.Second)
		}

		return false
	}
}

func (svc *service) stop() error {
	if svc.process == nil || svc.process.Process == nil {
		svc.logger.Warn("attempted to stop inactive service")
		return nil
	}

	svc.logger.Info("stopping service", "name", svc.config.Name)
	if err := svc.process.Process.Signal(syscall.SIGTERM); err != nil {
		svc.logger.Error("failed to send SIGTERM", "service", svc.config.Name, "err", err)
		return err
	}

	if err := svc.process.Wait(); err != nil {
		svc.logger.Error("failed to wait for process to stop", "service", svc.config.Name, "err", err)
		return err
	}

	return nil
}
