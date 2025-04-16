package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

type System struct {
	logger      *slog.Logger
	config      Config
	services    *ServiceManager
	filesystems *FilesystemManager
}

func NewSystem(logger *slog.Logger) (*System, error) {
	config, err := ReadConfig("/etc/orc.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &System{
		logger:      logger,
		config:      config,
		services:    NewServiceManager(logger, config.Services),
		filesystems: NewFilesystemManager(logger),
	}, nil
}

func (s *System) Start() error {
	if err := s.filesystems.MountAll(); err != nil {
		return fmt.Errorf("failed to mount filesystems: %w", err)
	}

	if err := s.services.StartAll(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	s.runShell()

	return nil
}

func (s *System) Stop() error {
	if err := s.services.StopAll(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	if err := s.filesystems.UnmountAll(); err != nil {
		return fmt.Errorf("failed to unmount filesystems: %w", err)
	}

	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		return fmt.Errorf("failed to power off: %w", err)
	}

	return nil
}

func (s *System) runShell() {
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
