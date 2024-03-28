#!/bin/sh
set -e

base=$(pwd)

build() {
    clean
    echo "building in cli"
    cd "$base/cli"
    go build
}

clean() {
    echo "cleaning in cli"
    cd "$base/cli"
    rm -f cli
}

test() {
    go test ./pkg/cmd $@
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
    test)
        test "$2"
        ;;
    *)
        echo 'usage: make [ clean | build ]'
        ;;
esac
