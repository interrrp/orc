#!/bin/sh

exe=$(mktemp)
go build -o "$exe" .
sudo unshare --fork --pid --mount-proc "$exe"
rm "$exe"
