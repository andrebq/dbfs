.PHONY: build tidy

build:
	go build ./...

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd