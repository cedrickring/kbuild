all: deps lint fmt vet test build

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

test:
	go test ./...

build-all:
	mkdir -p out && cd out
	which gox || make gox
	gox -arch="386 amd64" -os="darwin linux windows" --output "out/kbuild_{{.OS}}_{{.Arch}}" github.com/cedrickring/kbuild/cmd
