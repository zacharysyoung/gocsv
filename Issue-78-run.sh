#!/bin/zsh

set -e

literal=$(echo '\U005A')
echo $literal

gocsv replace -regex='z' -repl=$literal Issue-78-input.csv
