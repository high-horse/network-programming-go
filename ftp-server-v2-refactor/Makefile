build: 
	go build -o bin/simple/ftp *.go

build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/static/ftp *.go

build-fedora:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/fedora/ftp *.go