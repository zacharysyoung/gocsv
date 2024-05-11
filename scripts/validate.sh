#!/bin/sh
# Run go vet and go test for all code.
set -e

echo '-- validating --'
go vet  ./... && echo 'vet ok' || exit 1
go test ./...
