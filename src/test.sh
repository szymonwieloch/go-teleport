#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# client tests
cd $SCRIPT_DIR/client

echo CLIENT tests ===========================================
go test  -race -coverprofile cover.out -coverpkg=. ./...
go tool cover -func ./cover.out | grep total

# server tests
cd $SCRIPT_DIR/server

echo SERVER tests ===========================================
go test -race  -tags apitests -coverprofile cover.out -coverpkg=.,./jobs,./service ./...
go tool cover -func ./cover.out | grep total

