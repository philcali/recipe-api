#!/usr/bin/env bash

set -e

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main cmd/recipes/main.go
zip build_function.zip main