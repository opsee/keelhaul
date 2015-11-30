#!/bin/bash
set -e

echo "loading schema for tests..."
echo "create database keelhaul_test" | psql -U postgres -h postgresql
# echo "drop database if exists keelhaul_test; create database keelhaul_test" | psql -U postgres -h postgresql
migrate -url $POSTGRES_CONN -path ./migrations up
