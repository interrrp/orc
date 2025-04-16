package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/lmittmann/tint"
)

func main() {
	setUpLogging()

	if err := mountFilesystems(); err != nil {
		slog.Error("failed to mount filesystems", "err", err)
		os.Exit(1)
	}

	cfg, err := readConfig("/etc/orc.toml")
	if err != nil {
		slog.Error("failed to read config", "err", err)
		os.Exit(1)
	}

	if err := startServices(cfg.Services); err != nil {
		slog.Error("failed to start services", "err", err)
		os.Exit(1)
	}

	if err := runShell(); err != nil {
		slog.Error("shell exited with error", "err", err)
	}

	shutdown(cfg)
}

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

type service struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
	LogFile string `toml:"log_file"`
	process *exec.Cmd
}

type config struct {
	Services []service
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

func readConfig(path string) (config, error) {
	var cfg config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config from %s: %w", path, err)
	}
	return cfg, nil
}

func startServices(services []service) error {
	for i, svc := range services {
		slog.Info("starting service", "name", svc.Name, "command", svc.Command)

		process := exec.Command(svc.Command)
		services[i].process = process

		if svc.LogFile != "" {
			logFile, err := os.Create(svc.LogFile)
			if err != nil {
				return fmt.Errorf("failed to open log file %s for service %s: %w", svc.LogFile, svc.Name, err)
			}
			process.Stdout = logFile
			process.Stderr = logFile
		}

		if err := process.Start(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", svc.Name, err)
		}
	}
	return nil
}

func stopServices(services []service) error {
	var lastErr error
	for i := len(services) - 1; i >= 0; i-- {
		svc := services[i]
		if svc.process == nil || svc.process.Process == nil {
			continue
		}

		slog.Info("stopping service", "name", svc.Name)
		if err := svc.process.Process.Signal(syscall.SIGTERM); err != nil {
			slog.Error("failed to send SIGTERM", "service", svc.Name, "err", err)
			lastErr = err
			continue
		}

		if err := svc.process.Wait(); err != nil {
			slog.Error("failed to wait for process to stop", "service", svc.Name, "err", err)
			lastErr = err
		}
	}
	return lastErr
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
	if err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		slog.Error("failed to power off", "err", err)
		os.Exit(1)
	}
}
