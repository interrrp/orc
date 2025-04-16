package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

func mount(source, target, fsType string) {
	slog.Info("mounting", "source", source, "target", target, "type", fsType)

	if err := os.MkdirAll(target, 0755); err != nil {
		fatal("failed to create directory for mounting", "err", err)
	}

	if err := syscall.Mount(source, target, fsType, 0, ""); err != nil {
		fatal("failed to mount", "err", err)
	}
}

func main() {
	mount("proc", "/proc", "proc")
	mount("sys", "/sys", "sysfs")
	mount("tmpfs", "/run", "tmpfs")
	mount("udev", "/dev", "devtmpfs")
	mount("devpts", "/dev/pts", "devpts")

	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("failed to run shell", "err", err)
	}
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
