#!/usr/bin/env bash

set -e

ENTRYPOINT="${1:-"cmd/recipes/main.go"}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap "$ENTRYPOINT"
zip build_function.zip bootstrap