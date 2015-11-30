#!/bin/bash

docker pull sameersbn/postgresql:9.4-3
docker run --name postgresql -d -e PSQL_TRUST_LOCALNET=true -e DB_USER=postgres -e DB_PASS= -e DB_NAME=keelhaul_test sameersbn/postgresql:9.4-3
echo "started postgresql"
docker pull nsqio/nsq:latest
docker run --name lookupd -d nsqio/nsq /nsqlookupd
echo "started lookupd"
docker run --name nsqd --link lookupd:lookupd -d nsqio/nsq /nsqd --broadcast-address=nsqd --lookupd-tcp-address=lookupd:4160
echo "started nsqd"

DOCKER_GARBAGE=$(docker-machine ls -q | head -1)
HOST_IP=$(docker-machine ip $DOCKER_GARBAGE)

docker rm -f etcd || true
docker run -d -p 4001:4001 -p 2380:2380 -p 2379:2379 --name etcd quay.io/coreos/etcd:v2.0.3 \
 -name etcd0 \
 -advertise-client-urls http://${HOST_IP}:2379,http://${HOST_IP}:4001 \
 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
 -initial-advertise-peer-urls http://${HOST_IP}:2380 \
 -listen-peer-urls http://0.0.0.0:2380 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster etcd0=http://${HOST_IP}:2380 \
 -initial-cluster-state new
