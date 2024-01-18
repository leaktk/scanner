VERSION := 0.0.1
COMMIT := $(shell git rev-parse HEAD)
BUILD_META :=
BUILD_META += -X=github.com/leaktk/scanner/cmd.Version=$(VERSION)
BUILD_META += -X=github.com/leaktk/scanner/cmd.Commit=$(COMMIT)

LDFLAGS := -ldflags "$(BUILD_META)"

all: build

clean:
	if [[ -e .git ]]; then git clean -dfX; fi

build: format
	golint ./...
	go vet ./...
	go mod tidy
	go build $(LDFLAGS) -o leaktk-scanner

format:
	go fmt ./...

test: format
	go get golang.org/x/lint/golint
	go vet ./...
	golint ./...
	go test ./...
