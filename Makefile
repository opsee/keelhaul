APPENV ?= testenv

all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg proto/bastion_proto

fmt:
	@gofmt -w ./

deps:
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
		--env-file ./testenv \
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

deploy-cf:
	aws s3 cp --content-disposition inline --content-type application/json --region us-east-1 --acl public-read etc/bastion-cf.template s3://opsee-bastion-cf-us-east-1/beta/
	aws s3 cp --content-disposition inline --content-type application/json --region us-east-1 --acl public-read etc/bastion-ingress-cf.template s3://opsee-bastion-cf-us-east-1/beta/
	for region in ap-northeast-1 ap-northeast-2 ap-southeast-1 ap-southeast-2 eu-central-1 eu-west-1 sa-east-1 us-west-1 us-west-2; do \
		aws s3 cp --content-disposition inline --content-type application/json --source-region us-east-1 --region $$region --acl public-read s3://opsee-bastion-cf-us-east-1/beta/bastion-cf.template s3://opsee-bastion-cf-$$region/beta/ ; \
		aws s3 cp --content-disposition inline --content-type application/json --source-region us-east-1 --region $$region --acl public-read s3://opsee-bastion-cf-us-east-1/beta/bastion-ingress-cf.template s3://opsee-bastion-cf-$$region/beta/ ; \
	done

.PHONY: docker dbuild drun run migrate clean all
