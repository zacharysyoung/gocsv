#!/bin/sh
# Check git state, validate, build cmd/cli and install it 
# as ~/bin/csv.
set -e

[ -n "$(git status --porcelain)" ] && [ "$1" != "-F" ] && { 
    echo 'error: dirty git'
    exit 1
}

./scripts/validate.sh

echo '-- building --'
go build ./cmd/cli 
mv cli ~/bin/csv   && echo 'installed cli -> ~/bin/csv' || exit 1