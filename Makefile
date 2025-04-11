VERSION := $(shell ./hack/version)
COMMIT := $(shell git rev-parse HEAD)
BUILD_META :=
BUILD_META += -X=github.com/leaktk/scanner/cmd.Version=$(VERSION)
BUILD_META += -X=github.com/leaktk/scanner/cmd.Commit=$(COMMIT)
PREFIX ?= /usr
MODULE := $(shell grep '^module' go.mod | awk '{print $$2}')

SHELL := $(shell command -v bash;)
BASHINSTALLDIR=${PREFIX}/share/bash-completion/completions
ZSHINSTALLDIR=${PREFIX}/share/zsh/site-functions
FISHINSTALLDIR=${PREFIX}/share/fish/vendor_completions.d

SELINUXOPT ?= $(shell test -x /usr/sbin/selinuxenabled && selinuxenabled && echo -Z)

LDFLAGS := -ldflags "$(BUILD_META)"

all: build completions

clean:
	if [[ -e .git ]]; then git clean -dfX; fi

.PHONY: completions
completions: build
	declare -A outfiles=([bash]=%s [zsh]=_%s [fish]=%s.fish [powershell]=%s.ps1);\
	for shell in $${!outfiles[*]}; do \
		outfile=$$(printf "completions/$$shell/$${outfiles[$$shell]}" leaktk-scanner); \
		./leaktk-scanner completion $$shell >| $$outfile; \
	done

.PHONY: gosec
gosec:
	which gosec &> /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

.PHONY: lint
lint:
	podman run \
		--rm -v "$(PWD):/mnt:ro" \
		--security-opt label=disable \
		--workdir /mnt \
		ghcr.io/mgechev/revive:v1.3.7 \
			-formatter stylish ./...

build: format test
	go mod tidy
	go build $(LDFLAGS) -o leaktk-scanner

format:
	go fmt ./...
	which goimports &> /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -local $(MODULE) -l -w .

test: format gosec lint
	go vet ./...
	go test -race $(MODULE) ./...

install:
	install ./leaktk-scanner $(DESTDIR)$(PREFIX)/bin/leaktk-scanner

.PHONY: install.completions
install.completions:
	install ${SELINUXOPT} -d -m 755 $(DESTDIR)${BASHINSTALLDIR}
	install ${SELINUXOPT} -m 644 completions/bash/leaktk-scanner $(DESTDIR)${BASHINSTALLDIR}
	install ${SELINUXOPT} -d -m 755 $(DESTDIR)${ZSHINSTALLDIR}
	install ${SELINUXOPT} -m 644 completions/zsh/_leaktk-scanner $(DESTDIR)${ZSHINSTALLDIR}
	install ${SELINUXOPT} -d -m 755 $(DESTDIR)${FISHINSTALLDIR}
	install ${SELINUXOPT} -m 644 completions/fish/leaktk-scanner.fish $(DESTDIR)${FISHINSTALLDIR}

security-report:
	trivy fs .

update:
	go get -u ./...
	go mod tidy

.PHONY: validate.completions
validate.completions: SHELL:=/usr/bin/env bash
validate.completions: completions
	. completions/bash/leaktk-scanner
	if [ -x /bin/zsh ]; then /bin/zsh completions/zsh/_leaktk-scanner; fi
	if [ -x /bin/fish ]; then /bin/fish completions/fish/leaktk-scanner.fish; fi
