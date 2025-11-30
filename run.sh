#!/bin/bash
set -e

cleanup() {
    echo "Error occurred! Stopping docker-compose..."
    docker-compose down || true

    echo "Press ENTER to exit..."
    read
}

finish() {
    echo "Script finished successfully"
}

trap cleanup ERR

echo "=== Step 1: Build application ==="
docker-compose build

echo "=== Step 2: Run integration tests ==="
docker-compose up -d
docker build -f tests/Dockerfile -t integration-tests .
docker-compose down
trap - ERR

echo "=== Step 3: Run application ==="
docker-compose up -d

finish

echo "Press ENTER to exit..."
read
