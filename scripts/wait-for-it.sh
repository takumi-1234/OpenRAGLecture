#!/bin/sh
# wait-for-it.sh: Wait for a TCP host/port to be available before executing a command.
#
# Author: Giles Hall
#
# This script is a simplified version of the full script available at:
# https://github.com/vishnubob/wait-for-it
# It has been adapted to be POSIX-compliant and use minimal dependencies.

set -e

TIMEOUT=15
QUIET=0
HOST=""
PORT=""

usage() {
    echo "Usage: $0 host:port [-t timeout] [-- command args...]"
    echo "  -t TIMEOUT | --timeout=TIMEOUT  Timeout in seconds, default is 15"
    echo "  -q | --quiet                   Don't output any status messages"
    echo "  -- COMMAND ARGS...              Execute command with args after the test finishes"
}

wait_for() {
    if [ $QUIET -eq 0 ]; then echo "Waiting for $HOST:$PORT..."; fi
    for i in `seq $TIMEOUT`; do
        if nc -z "$HOST" "$PORT" >/dev/null 2>&1; then
            if [ $QUIET -eq 0 ]; then echo "$HOST:$PORT is available after $i seconds"; fi
            return 0
        fi
        sleep 1
    done
    echo "Timeout occurred after waiting $TIMEOUT seconds for $HOST:$PORT"
    return 1
}

while [ $# -gt 0 ]
do
    case "$1" in
        *:* )
        HOST=$(printf "%s\n" "$1"| cut -d : -f 1)
        PORT=$(printf "%s\n" "$1"| cut -d : -f 2)
        shift 1
        ;;
        -q | --quiet)
        QUIET=1
        shift 1
        ;;
        -t)
        TIMEOUT="$2"
        if [ "$TIMEOUT" = "" ]; then break; fi
        shift 2
        ;;
        --timeout=*)
        TIMEOUT="${1#*=}"
        shift 1
        ;;
        --)
        shift
        CMD="$@"
        break
        ;;
        -h | --help)
        usage
        exit 0
        ;;
        *)
        usage
        exit 1
        ;;
    esac
done

if [ "$HOST" = "" ] || [ "$PORT" = "" ]; then
    echo "Error: you need to provide a host and port to test."
    usage
    exit 1
fi

wait_for

if [ "$CMD" != "" ]; then
    exec $CMD
fi