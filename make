#!/bin/sh
set -e

base=$(pwd)

clean() {
    echo "cleaning in cli"
    cd "$base/cli"
    rm -f cli
}

build() {
    clean
    echo "building in cli"
    cd "$base/cli"
    go build
}

case $1 in
    '')
        build
        ;;
    clean)
        clean
        ;;
    build)
        build
        ;;
    *)
        echo 'usage: make [ clean | build ]'
        ;;
esac
