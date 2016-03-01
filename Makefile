all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg proto/bastion_proto

fmt:
	@gofmt -w ./

deps:
	docker-compose stop
	docker-compose rm -f
	docker-compose up -d
	docker run --link keelhaul_postgresql:postgres aanand/wait

migrate:
	migrate -url $(KEELHAUL_POSTGRES_CONN) -path ./migrations up

docker: deps fmt dbuild

run: docker drun

dbuild:
	docker run \
		--link keelhaul_postgresql:postgresql \
		--link keelhaul_nsqd:nsqd \
		--link keelhaul_lookupd:lookupd \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-e GODEBUG=netdns=cgo \
		-v `pwd`:/build quay.io/opsee/build-go:go15 \
		&& docker build -t quay.io/opsee/keelhaul .

drun:
	docker run \
		--link keelhaul_postgresql:postgresql \
		--link keelhaul_nsqd:nsqd \
		--link keelhaul_lookupd:lookupd \
		--link keelhaul_etcd:etcd \
		--env-file ./$(APPENV) \
		-e GODEBUG=netdns=cgo \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-p 9094:9094 \
		--rm \
		quay.io/opsee/keelhaul:latest

.PHONY: docker dbuild drun run migrate clean all
