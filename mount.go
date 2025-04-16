package main

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"
)

func mount(source, target, fsType string) error {
	slog.Info("mounting", "source", source, "target", target, "type", fsType)

	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create directory for mounting %s: %w", target, err)
	}

	if err := syscall.Mount(source, target, fsType, 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s to %s: %w", source, target, err)
	}

	return nil
}

func unmount(target string) error {
	slog.Info("unmounting", "target", target)
	if err := syscall.Unmount(target, 0); err != nil {
		return fmt.Errorf("failed to unmount %s: %w", target, err)
	}
	return nil
}

func mountFilesystems() error {
	filesystems := []struct {
		source string
		target string
		fsType string
	}{
		{"proc", "/proc", "proc"},
		{"sys", "/sys", "sysfs"},
		{"tmpfs", "/run", "tmpfs"},
		{"udev", "/dev", "devtmpfs"},
		{"devpts", "/dev/pts", "devpts"},
	}

	for _, fs := range filesystems {
		if err := mount(fs.source, fs.target, fs.fsType); err != nil {
			return err
		}
	}
	return nil
}

func unmountFilesystems() error {
	filesystems := []string{
		"/dev/pts",
		"/run",
		"/sys",
		"/proc",
	}

	var lastErr error
	for _, fs := range filesystems {
		if err := unmount(fs); err != nil {
			slog.Error("failed to unmount", "target", fs, "err", err)
			lastErr = err
		}
	}
	return lastErr
}
