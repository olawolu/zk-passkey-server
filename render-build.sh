#!/usr/bin/env bash
# exit on error
set -o errexit

git config --global url."https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
go build -tags netgo -ldflags '-s -w' -o passkey-server cmd/server/main.go