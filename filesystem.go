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
	Source string
	Target string
	Type   string
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
		if fs.Target == "/dev" {
			// Unmounting /dev will cause issues
			return
		}

		if err := fm.unmount(fs.Target); err != nil {
			lastErr = err
		}
	})
	return lastErr
}

func (fm *FilesystemManager) mount(fs filesystem) error {
	fm.logger.Info("mounting", "source", fs.Source, "target", fs.Target, "type", fs.Type)

	if err := os.MkdirAll(fs.Target, 0755); err != nil {
		return fmt.Errorf("failed to create directory for mounting %s: %w", fs.Target, err)
	}

	if err := syscall.Mount(fs.Source, fs.Target, fs.Type, 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s to %s: %w", fs.Source, fs.Target, err)
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
