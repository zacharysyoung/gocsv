#!/bin/sh
# Validate the tgt directory and all its descendants.
set -e

tgt="$1"
[ -z "$tgt" ] && tgt='.'
cd "$tgt"

echo '-- validating --'

grep -nr 'fmt\.Print' ./subcmd            && exit 1 # catch debug-Print* statements, don't care about Sprint* or Fprint*

gofmt      -w  .     && echo 'fmt ok'     || exit 1
goimports  -w  .     && echo 'imports ok' || exit 1
go vet         ./... && echo 'vet ok'     || exit 1
go test        ./...

