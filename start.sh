#!/bin/sh

exe=$(mktemp)
go build -o "$exe" .
sudo unshare --fork --pid "$exe"
rm "$exe"
