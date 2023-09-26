#!/usr/bin/env bash

set -e

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap cmd/recipes/main.go
zip build_function.zip bootstrap