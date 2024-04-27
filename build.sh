#!/bin/sh

[ -z "$1" ] && { echo 'usage: build GOCSV-DIR'; exit 1; }

cd "$1"
go build ./cmd/cli
mv cli ~/bin/csv