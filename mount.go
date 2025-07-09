package main

import (
	"log/slog"
	"os"
	"syscall"
)

func mountAllFilesystems() {
	mount("proc", "/proc", "proc")
	mount("sys", "/sys", "sysfs")
	mount("tmpfs", "/run", "tmpfs")
	mount("udev", "/dev", "devtmpfs")
	mount("devpts", "/dev/pts", "devpts")
}

func unmountAllFileSystems() {
	unmount("/dev/pts")
	unmount("/run")
	unmount("/sys")
	unmount("/proc")
}

func mount(source, target, fsType string) {
	slog.Info("mounting", "target", target)

	if err := os.MkdirAll(target, 0600); err != nil {
		slog.Error("failed to create directory for mounting", "target", target)
		return
	}

	if err := syscall.Mount(source, target, fsType, 0, ""); err != nil {
		slog.Error("failed to mount",
			"err", err,
			"source", source,
			"target", target,
			"fsType", fsType)
	}
}

func unmount(target string) {
	slog.Info("unmounting", "target", target)

	if err := syscall.Unmount(target, 0); err != nil {
		slog.Error("failed to unmount", "err", err, "target", target)
	}
}
