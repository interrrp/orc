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

func mustMount(source, target, fsType string) {
	slog.Info("mounting", "source", source, "target", target, "type", fsType)

	must(os.MkdirAll(target, 0755),
		"failed to create directory for mounting")

	must(syscall.Mount(source, target, fsType, 0, ""),
		"failed to mount")
}

func mustUnmount(target string) {
	slog.Info("unmounting", "target", target)
	must(syscall.Unmount(target, 0), "failed to unmount")
}

type service struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
	LogFile string `toml:"log_file"`
	process *exec.Cmd
}

var cfg struct {
	Services []service
}

func main() {
	setUpLogging()

	mustMount("proc", "/proc", "proc")
	mustMount("sys", "/sys", "sysfs")
	mustMount("tmpfs", "/run", "tmpfs")
	mustMount("udev", "/dev", "devtmpfs")
	mustMount("devpts", "/dev/pts", "devpts")

	_, err := toml.DecodeFile("/etc/orc.toml", &cfg)
	must(err, "failed to read config")

	startServices()

	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run(), "failed to run shell")

	shutdown()
}

func shutdown() {
	mustUnmount("/proc")
	mustUnmount("/sys")
	mustUnmount("/run")
	mustUnmount("/dev/pts")

	stopServices()

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

func must(err error, msg string, args ...any) {
	if err != nil {
		args = append(args, "err", err)
		slog.Error(msg, args...)
		os.Exit(1)
	}
}
