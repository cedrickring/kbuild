all: deps lint fmt vet build

deps:
	go get github.com/golang/lint/golint

build:
	mkdir -p bin
	go build -o bin/kbuild ./cmd/kbuild.go

fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

lint:
	golint ./pkg/... ./cmd/...

gox:
	go get github.com/mitchellh/gox
	gox -build-toolchain

build-all:
	which gox || make gox
	gox -arch="386 amd64" -os="darwin linux windows" github.com/cedrickring/kbuild/cmd/kbuild