.PHONY: build tidy watch dist

OS?=UnixFamily

include Env-$(OS).mk

test:
	go build ./...

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd

dist:
	mkdir -p $(DIST_FOLDER)
	go build -o $(DIST_BINARY_FILE) ./cmd/dbfs
