all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg proto/bastion_proto

fmt:
	@gofmt -w ./

deps:
	./deps.sh

migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

docker: fmt
	docker run \
		--link postgresql:postgresql \
		--link nsqd:nsqd \
		--link lookupd:lookupd \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-e GODEBUG=netdns=cgo \
		-v `pwd`:/build quay.io/opsee/build-go:go15 \
		&& docker build -t quay.io/opsee/keelhaul .

run: docker
	docker run \
		--link postgresql:postgresql \
		--link nsqd:nsqd \
		--link lookupd:lookupd \
		--link etcd:etcd \
		--env-file ./$(APPENV) \
		-e GODEBUG=netdns=cgo \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-p 9094:9094 \
		--rm \
		quay.io/opsee/keelhaul:latest

.PHONY: docker run migrate clean all
