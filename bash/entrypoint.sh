#!/usr/bin/env bash
set -eou pipefail

function run-cmd {
    declare -r cmd=${1}
    shift
    "${cmd}" "${@}"
}

case $# in
    0) /usr/local/bin/webdav
        ;;
    1) "${1}"
        ;;
    *) run-cmd "${@}"
esac
