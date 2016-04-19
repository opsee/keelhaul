#!/bin/bash
set -e

echo "loading schema for tests..."
echo "create database keelhaul_test" | psql $KEELHAUL_POSTGRES_CONN
#echo "drop database if exists keelhaul_test; create database keelhaul_test" | psql $KEELHAUL_POSTGRES_CONN
migrate -url $KEELHAUL_POSTGRES_CONN -path ./migrations up

go run generate.go
