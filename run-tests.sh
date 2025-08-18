#!/bin/sh
set -e

echo "==== Running Unit Tests ===="
go test -v ./...

echo "==== Running E2E Tests ===="
until nc -z clickhouse-test 9000; do
  echo "Waiting for ClickHouse..."
  sleep 1
done

go test -v ./tests/e2e/... --tags=e2e
