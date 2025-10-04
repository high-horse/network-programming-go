build: 
	go build -o bin/simple/fingerprint sample/*.go

build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ftp *.go

build-fedora:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/fedora/ftp *.go