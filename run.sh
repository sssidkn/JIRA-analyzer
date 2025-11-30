set -e

echo "=== Step 1: Build application ==="
docker-compose build

echo "=== Step 2: Run unit tests ==="
cd JIRA-connector
go test -v ./...
cd ..

echo "=== Step 3: Run integration tests ==="
cd tests/integration
go test -v ./...
cd ../..

echo "=== Step 4: Run application ==="
docker-compose up
