.PHONY: build tidy

build:
	go build ./...

dist:
	go build -o dist/webdav ./cmd/webdav
	go build -o dist/dbfs ./cmd/dbfs

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd

docker:
	docker build -t andrebq/dbfs:latest .
