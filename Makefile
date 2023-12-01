all: build

clean:
	if [[ -e .git ]]; then git clean -dfX; fi

build:
	golint ./...
	go vet ./...
	go mod tidy
	go build -o ./leaktk-scanner
