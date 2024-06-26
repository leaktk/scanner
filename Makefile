VERSION := 0.0.1
COMMIT := $(shell git rev-parse HEAD)
BUILD_META :=
BUILD_META += -X=github.com/leaktk/scanner/cmd.Version=$(VERSION)
BUILD_META += -X=github.com/leaktk/scanner/cmd.Commit=$(COMMIT)
PREFIX := /usr

LDFLAGS := -ldflags "$(BUILD_META)"

all: build

clean:
	if [[ -e .git ]]; then git clean -dfX; fi

.PHONY: gosec
gosec:
	which gosec &> /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

.PHONY: golint
golint:
	which golint &> /dev/null || go install golang.org/x/lint/golint@latest
	golint ./...

build: format test
	go mod tidy
	go build $(LDFLAGS) -o leaktk-scanner

format:
	go fmt ./...

test: format gosec golint
	go vet ./...
	go test ./...

install: build
	install ./leaktk-scanner $(DESTDIR)$(PREFIX)/bin/leaktk-scanner

security-report:
	trivy fs .

update:
	go get -u ./...
	go mod tidy
