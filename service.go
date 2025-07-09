package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func loadServices() []*service {
	serviceDir := "/etc/orc"

	var services []*service

	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		slog.Error("failed to read service directory", "err", err, "dir", serviceDir)
		return services
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			services = append(services, &service{
				name:       entry.Name(),
				scriptPath: filepath.Join(serviceDir, entry.Name()),
			})
		}
	}

	return services
}

func startAllServices(services []*service) {
	for _, s := range services {
		s.start()
	}
}

func stopAllServices(services []*service) {
	for i := len(services) - 1; i >= 0; i-- {
		services[i].stop()
	}
}

type service struct {
	name       string
	scriptPath string
	process    *exec.Cmd
}

func (svc *service) start() {
	slog.Info("starting service", "name", svc.name)

	cmd := exec.Command("/bin/sh", "-c", ". "+svc.scriptPath+" && start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		slog.Error("failed to start service", "err", err, "name", svc.name)
	}
	svc.process = cmd
}

func (svc *service) stop() {
	if svc.process == nil || svc.process.Process == nil {
		return
	}

	slog.Info("stopping service", "name", svc.name)

	stopCmd := exec.Command("/bin/sh", "-c", ". "+svc.scriptPath+" && stop")
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr

	if err := stopCmd.Run(); err != nil {
		slog.Warn("failed to run stop function", "err", err)
	}

	if err := svc.process.Process.Signal(syscall.SIGTERM); err != nil {
		slog.Warn("failed to send SIGTERM, killing process", "err", err)
		if err := svc.process.Process.Kill(); err != nil {
			slog.Warn("failed to kill process", "err", err)
		}
	}

	if err := svc.process.Wait(); err != nil {
		slog.Warn("failed to wait for process to exit", "err", err)
	}

	svc.process = nil
}
