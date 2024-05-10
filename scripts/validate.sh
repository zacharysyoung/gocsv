#!/bin/sh
# Run go vet and go test for all code.
set -e

go vet  ./... && echo 'ok	vet'
go test ./...
