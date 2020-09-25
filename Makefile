.DEFAULT_GOAL := binary

GO := go

binary: bin/norouter

bin/norouter:
	CGO_ENABLED=0 $(GO) build -o $@ ./cmd/norouter
	if [ $(shell go env GOOS) = linux ]; then LANG=C LC_ALL=C file $@ | grep -qw "statically linked"; fi

# The file name convention for Unix: ./bin/norouter-$(uname -s)-$(uname -m)
cross:
	CGO_ENABLED=0 GOOS=linux     GOARCH=amd64 $(GO) build -o ./bin/norouter-Linux-x86_64     ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=arm64 $(GO) build -o ./bin/norouter-Linux-aarch64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=darwin    GOARCH=amd64 $(GO) build -o ./bin/norouter-Darwin-x86_64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=freebsd   GOARCH=amd64 $(GO) build -o ./bin/norouter-FreeBSD-amd64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=netbsd    GOARCH=amd64 $(GO) build -o ./bin/norouter-NetBSD-amd64     ./cmd/norouter
	CGO_ENABLED=0 GOOS=openbsd   GOARCH=amd64 $(GO) build -o ./bin/norouter-OpenBSD-amd64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=dragonfly GOARCH=amd64 $(GO) build -o ./bin/norouter-DragonFly-x86_64 ./cmd/norouter
	CGO_ENABLED=0 GOOS=windows   GOARCH=amd64 $(GO) build -o ./bin/norouter-Windows-x64.exe  ./cmd/norouter

clean:
	rm -rf bin

integration:
	./integration/test-internal-agent.sh
	./integration/test-router.sh

.PHONY: binary bin/norouter cross clean integration
