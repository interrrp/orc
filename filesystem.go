package main

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"
)

type FilesystemManager struct {
	logger      *slog.Logger
	filesystems []filesystem
}

type filesystem struct {
	source string
	target string
	fsType string
}

func NewFilesystemManager(logger *slog.Logger) *FilesystemManager {
	return &FilesystemManager{
		logger: logger,
		filesystems: []filesystem{
			{"proc", "/proc", "proc"},
			{"sys", "/sys", "sysfs"},
			{"tmpfs", "/run", "tmpfs"},
			{"udev", "/dev", "devtmpfs"},
			{"devpts", "/dev/pts", "devpts"},
		},
	}
}

func (fm *FilesystemManager) MountAll() error {
	for _, fs := range fm.filesystems {
		if err := fm.mount(fs); err != nil {
			return err
		}
	}
	return nil
}

func (fm *FilesystemManager) UnmountAll() error {
	var lastErr error
	iterateReverse(fm.filesystems, func(fs filesystem) {
		if fs.target == "/dev" {
			// Unmounting /dev will cause issues
			return
		}

		if err := fm.unmount(fs.target); err != nil {
			lastErr = err
		}
	})
	return lastErr
}

func (fm *FilesystemManager) mount(fs filesystem) error {
	fm.logger.Info("mounting", "source", fs.source, "target", fs.target, "type", fs.fsType)

	if err := os.MkdirAll(fs.target, 0755); err != nil {
		return fmt.Errorf("failed to create directory for mounting %s: %w", fs.target, err)
	}

	if err := syscall.Mount(fs.source, fs.target, fs.fsType, 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s to %s: %w", fs.source, fs.target, err)
	}

	return nil
}

func (fm *FilesystemManager) unmount(target string) error {
	fm.logger.Info("unmounting", "target", target)
	if err := syscall.Unmount(target, 0); err != nil {
		return fmt.Errorf("failed to unmount %s: %w", target, err)
	}
	return nil
}
