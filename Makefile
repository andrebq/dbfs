.PHONY: build tidy watch dist

OS?=UnixFamily

include Env-Common.mk
include Env-$(OS).mk
include SupportApps.mk

test:
	go test ./...

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd

dist:
	go build -o $(DIST_BINARY_FILE) ./cmd/dbfs
