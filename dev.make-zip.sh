#!/usr/bin/env bash

set -e

ENTRYPOINT="${1:-"cmd/recipes/main.go"}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap "$ENTRYPOINT"
FUNC="$(basename $(dirname "$ENTRYPOINT"))"
zip build_${FUNC}_function.zip bootstrap