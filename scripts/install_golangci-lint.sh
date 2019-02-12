#!/usr/bin/env bash

if ! [[ -x "$(command -v golangci-lint)" ]]; then
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.14.0
fi