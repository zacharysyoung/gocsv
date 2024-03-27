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
    run=$1
    path='./pkg/cmd'

    if [ -z "$run" ]; then
        go test $path
    else
        go test -run "$run" $path
    fi
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
