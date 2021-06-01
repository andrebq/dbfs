.PHONY: build tidy watch dist test test-minio ll-tests

OS?=UnixFamily

include Env-Common.mk
include Env-$(OS).mk
include SupportApps.mk
include LocalDevTasks.mk

test:
	go test ./...

test-minio:
	AWS_ACCESS_KEY_ID=dbfs-dev \
	AWS_SECRET_ACCESS_KEY=dbfs-dev \
	DBFS_MINIO_ENDPOINT=localhost:9000 \
	DBFS_BUCKET=dbfs-data \
	AWS_DEFAULT_REGION=us-east-1 \
		go test --tags minio_driver ./...

all-tests: test test-minio

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd

dist:
	go build -o $(DIST_BINARY_FILE) ./cmd/dbfs
