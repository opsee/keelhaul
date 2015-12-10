#!/bin/bash
set -e

echo "loading schema for tests..."
echo "create database keelhaul_test" | psql -U postgres -h postgresql
# echo "drop database if exists keelhaul_test; create database keelhaul_test" | psql -U postgres -h postgresql
migrate -url $POSTGRES_CONN -path ./migrations up

proto_dir=proto/bastion_proto
checker_proto=${proto_dir}/checker.proto
protoc -I/usr/local/include -I${proto_dir}/ --go_out=plugins=grpc:${proto_dir}/ ${checker_proto}
rm -f src/github.com/opsee/bastion/checker/checker.pb.go
ln -s ../../../../../proto/checker.pb.go src/github.com/opsee/keelhaul/checker/checker.pb.go

if [ "$GOOS" = "darwin" ]; then
  sed -i'' -e 's/import google_protobuf "google\/protobuf"/import google_protobuf "go.pedge.io\/google-protobuf"/' ${proto_dir}/checker.pb.go
else
  sed -i 's/import google_protobuf "google\/protobuf"/import google_protobuf "go.pedge.io\/google-protobuf"/' ${proto_dir}/checker.pb.go
fi
