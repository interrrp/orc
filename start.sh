#!/bin/sh
set -eu

ORIG_PWD=$(pwd)
ROOTFS_URL="https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz"
VM="$ORIG_PWD/vm"
ROOTFS="$VM/rootfs"
ROOTFS_INIT="$ROOTFS/sbin/init"
ROOTFS_IMAGE="$VM/initrd.img"
KERNEL=$(ls /boot/vmlinuz* | head -n 1)

build_orc() {
    go build -o "$VM/init" .
}

fetch_rootfs() {
    if [ -n "$(ls -A "$ROOTFS" 2>/dev/null)" ]; then
        echo "Rootfs already populated, skipping fetch."
        return
    fi

    mkdir -p "$ROOTFS"
    TMP_TAR="$VM/alpine-minirootfs.tar.gz"
    wget -O "$TMP_TAR" "$ROOTFS_URL"
    tar -xzf "$TMP_TAR" -C "$ROOTFS"
    rm "$TMP_TAR"
}

replace_init() {
    mv "$VM/init" "$ROOTFS_INIT"
}

build_rootfs_image() {
    cd "$ROOTFS"
    find . | cpio -H newc -o | gzip -9 > "$ROOTFS_IMAGE"
    cd "$ORIG_PWD"
}

run_vm() {
    qemu-system-x86_64 \
        -kernel "$KERNEL" \
        -initrd "$ROOTFS_IMAGE" \
        -append "console=ttyS0 rdinit=/sbin/init" \
        -nographic
}

build_orc
fetch_rootfs
replace_init
build_rootfs_image
run_vm
