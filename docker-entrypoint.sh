#!/bin/sh
set -e

# 명령어 처리
case "$1" in
    server)
        echo "Starting Chuingho Server..."
        exec ./server
        ;;
    migration)
        echo "Running database migration..."
        exec ./migration "$@"
        ;;
    prepare_phrases)
        echo "Preparing phrase candidates..."
        exec ./prepare_phrases "$@"
        ;;
    *)
        echo "Available commands: server, migration, prepare_phrases"
        echo "Running: $@"
        exec "$@"
        ;;
esac