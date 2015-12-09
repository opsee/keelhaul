#!/bin/bash
set -e

echo "loading schema for tests..."
echo "create database keelhaul_test" | psql -U postgres -h postgresql
#echo "drop database if exists keelhaul_test; create database keelhaul_test" | psql -U postgres -h postgresql
migrate -url $POSTGRES_CONN -path ./migrations up

git submodule update --init

proto_dir=proto/bastion_proto
checker_proto=${proto_dir}/checker.proto
protoc --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. ${checker_proto}

mv ./proto/bastion_proto/checker.pb.go src/github.com/opsee/keelhaul/checker/checker.pb.go
