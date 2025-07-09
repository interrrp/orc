package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	initLogger()

	mount("proc", "/proc", "proc")
	mount("sys", "/sys", "sysfs")
	mount("tmpfs", "/run", "tmpfs")
	mount("udev", "/dev", "devtmpfs")
	mount("devpts", "/dev/pts", "devpts")

	services := loadServices()
	for _, s := range services {
		s.start()
	}

	startShell()

	for i := len(services) - 1; i >= 0; i-- {
		services[i].stop()
	}

	unmount("/dev/pts")
	unmount("/run")
	unmount("/sys")
	unmount("/proc")

	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		slog.Error("failed to power off", "err", err)
	}
}

func startShell() {
	sh := exec.Command("/bin/sh")

	sh.Stdin = os.Stdin
	sh.Stdout = os.Stdout
	sh.Stderr = os.Stderr

	if err := sh.Start(); err != nil {
		slog.Error("failed to start shell", "err", err)
	}
	if err := sh.Wait(); err != nil {
		slog.Error("shell exited", "err", err)
	}
}

func initLogger() {
	opts := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
	}
	handler := tint.NewHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
