#!/bin/sh
set -e # exit-on-error

[ -z "$1" ] && { echo 'usage: build GOCSV-DIR'; exit 1; }

cd "$1"
go generate ./pkg/subcmd
go test ./pkg/subcmd/...
go test ./cmd/cli/...
go build ./cmd/cli
mv cli ~/bin/csv