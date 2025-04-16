package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

func mustMount(source, target, fsType string) {
	slog.Info("mounting", "source", source, "target", target, "type", fsType)

	if err := os.MkdirAll(target, 0755); err != nil {
		fatal("failed to create directory for mounting", "err", err)
	}

	if err := syscall.Mount(source, target, fsType, 0, ""); err != nil {
		fatal("failed to mount", "err", err)
	}
}

func mustUnmount(target string) {
	slog.Info("unmounting", "target", target)

	if err := syscall.Unmount(target, 0); err != nil {
		slog.Error("failed to unmount", "err", err)
	}
}

func main() {
	setUpLogging()

	mustMount("proc", "/proc", "proc")
	mustMount("sys", "/sys", "sysfs")
	mustMount("tmpfs", "/run", "tmpfs")
	mustMount("udev", "/dev", "devtmpfs")
	mustMount("devpts", "/dev/pts", "devpts")

	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fatal("failed to run shell", "err", err)
	}

	shutdown()
}

func shutdown() {
	mustUnmount("/proc")
	mustUnmount("/sys")
	mustUnmount("/run")
	mustUnmount("/dev/pts")

	slog.Info("powering off")
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		slog.Error("failed to power off", "err", err)
		slog.Error("init will be killed, expect a kernel panic")
		os.Exit(1)
	}
}

func setUpLogging() {
	opts := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
	}
	handler := tint.NewHandler(os.Stderr, opts)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	clearScreen()
}

func clearScreen() {
	fmt.Print("\033c")
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
