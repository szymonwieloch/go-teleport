#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# client tests
cd $SCRIPT_DIR/client
go test -v -race -coverprofile cover.out ./...

# server tests
cd $SCRIPT_DIR/server

echo $PWD

# API tests + unit tests
go test -v -race  -tags apitests -coverprofile cover.out ./... 

