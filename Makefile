.DEFAULT_GOAL := binary

GO := go

binary: bin/norouter

install:
	cp -f bin/norouter /usr/local/bin/norouter

uninstall:
	rm -f /usr/local/bin/norouter

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
	CGO_ENABLED=0 GOOS=openbsd   GOARCH=arm64 $(GO) build -o ./bin/norouter-OpenBSD-arm64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=dragonfly GOARCH=amd64 $(GO) build -o ./bin/norouter-DragonFly-x86_64 ./cmd/norouter
	CGO_ENABLED=0 GOOS=windows   GOARCH=amd64 $(GO) build -o ./bin/norouter-Windows-x64.exe  ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=arm64 $(GO) build -o ./bin/norouter-Linux-aarch64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64 $(GO) build -o ./bin/norouter-Linux-ppc64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64le $(GO) build -o ./bin/norouter-Linux-ppc64le    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=mips64 $(GO) build -o ./bin/norouter-Linux-mips64    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=mips64le $(GO) build -o ./bin/norouter-Linux-mips64le    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=s390x $(GO) build -o ./bin/norouter-Linux-s390x    ./cmd/norouter
	CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64le $(GO) build -o ./bin/norouter-Linux-ppc64le    ./cmd/norouter
#       CGO_ENABLED=0 GOOS=linux     GOARM=5 GOARCH=arm $(GO) build -o ./bin/norouter-Linux-armel    ./cmd/norouter
#	CGO_ENABLED=0 GOOS=linux     GOARM=5 GOARCH=arm $(GO) build -o ./bin/norouter-Linux-armv5l    ./cmd/norouter
#       CGO_ENABLED=0 GOOS=linux     GOARM=6 GOARCH=arm $(GO) build -o ./bin/norouter-Linux-armhf    ./cmd/norouter
#	CGO_ENABLED=0 GOOS=linux     GOARM=6 GOARCH=arm $(GO) build -o ./bin/norouter-Linux-armv6l    ./cmd/norouter
#	CGO_ENABLED=0 GOOS=linux     GOARM=7 GOARCH=arm $(GO) build -o ./bin/norouter-Linux-armv7l    ./cmd/norouter
#	CGO_ENABLED=0 GOOS=windows   GO386=387 GOARCH=386 $(GO) build -o ./bin/norouter-Windows-x32.exe  ./cmd/norouter
#	CGO_ENABLED=0 GOOS=js	     GOARCH=wasm $(GO) build -o ./bin/norouter-wasm  ./cmd/norouter


clean:
	rm -rf bin

integration:
	./integration/test-agent.sh
	./integration/test-integration.sh

.PHONY: binary install uninstall bin/norouter cross clean integration
