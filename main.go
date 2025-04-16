package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	setUpLogging()

	cfg, err := readConfig("/etc/orc.toml")
	must(err, "failed to read config")

	must(mountFilesystems(), "failed to mount filesystems")
	must(startServices(cfg.Services), "failed to start services")
	must(runShell(), "shell exited with error")

	shutdown(cfg)
}

func must(err error, msg string, args ...any) {
	if err != nil {
		args = append(args, "err", err)
		slog.Error(msg, args...)
		os.Exit(1)
	}
}

func runShell() error {
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shutdown(cfg config) {
	slog.Info("initiating shutdown")

	if err := stopServices(cfg.Services); err != nil {
		slog.Error("errors occurred while stopping services", "err", err)
	}

	if err := unmountFilesystems(); err != nil {
		slog.Error("errors occurred while unmounting filesystems", "err", err)
	}

	slog.Info("powering off")
	must(syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF), "failed to power off")
}
