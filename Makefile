.PHONY: build tidy

build:
	go build ./...

test: build
	go test ./...

dist: build
	go build -o dist/webdav ./cmd/webdav
	go build -o dist/dbfs ./cmd/dbfs
	go build -o dist/authfs-adduser ./cmd/authfs-adduser
	cp bash/entrypoint.sh dist/entrypoint.sh

dist-win: test
	go build -o dist/webdav.exe ./cmd/webdav
	go build -o dist/dbfs.exe ./cmd/dbfs
	go build -o dist/authfs-adduser.exe ./cmd/authfs-adduser

tidy: build
	go fmt ./...
	go mod tidy

watch:
	modd

docker:
	docker build -t andrebq/dbfs:latest .
