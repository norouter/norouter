.DEFAULT_GOAL := binary

GO := go

GO_BUILD := GO111MODULE=on CGO_ENABLED=0 $(GO) build -ldflags="-s -w"

TARFLAGS := --owner=0 --group=0

binary: bin/norouter

install:
	cp -f bin/norouter /usr/local/bin/norouter

uninstall:
	rm -f /usr/local/bin/norouter

bin/norouter:
	rm -f $@$(shell go env GOEXE)
	$(GO_BUILD) -o $@$(shell go env GOEXE) ./cmd/norouter
	if [ $(shell go env GOOS) = linux ]; then LANG=C LC_ALL=C file $@ | grep -qw "statically linked"; fi

# The file name convention for Unix: ./artifacts/norouter-$(uname -s)-$(uname -m).tgz
# The file name convention has changed in v0.6.0.
cross:
	rm -rf bin artifacts
	mkdir -p artifacts

	GOOS=linux     GOARCH=amd64 make binary
	tar czvf artifacts/norouter-Linux-x86_64.tgz     $(TARFLAGS) -C bin norouter

	GOOS=linux     GOARCH=arm64 make binary
	tar czvf artifacts/norouter-Linux-aarch64.tgz    $(TARFLAGS) -C bin norouter

	GOOS=darwin    GOARCH=amd64 make binary
	tar czvf artifacts/norouter-Darwin-x86_64.tgz    $(TARFLAGS) -C bin norouter

	GOOS=darwin    GOARCH=arm64 make binary
	tar czvf artifacts/norouter-Darwin-arm64.tgz     $(TARFLAGS) -C bin norouter

	GOOS=freebsd   GOARCH=amd64 make binary
	tar czvf artifacts/norouter-FreeBSD-amd64.tgz    $(TARFLAGS) -C bin norouter

	GOOS=netbsd    GOARCH=amd64 make binary
	tar czvf artifacts/norouter-NetBSD-amd64.tgz     $(TARFLAGS) -C bin norouter

	GOOS=openbsd   GOARCH=amd64 make binary
	tar czvf artifacts/norouter-OpenBSD-amd64.tgz    $(TARFLAGS) -C bin norouter

	GOOS=dragonfly GOARCH=amd64 make binary
	tar czvf artifacts/norouter-DragonFly-x86_64.tgz $(TARFLAGS) -C bin norouter

	GOOS=windows   GOARCH=amd64 make binary
	zip -X -j artifacts/norouter-Windows-x64.zip bin/norouter.exe 

	rm -rf bin

clean:
	rm -rf bin artifacts

integration:
	./integration/test-agent.sh
	./integration/test-integration.sh

.PHONY: binary install uninstall bin/norouter cross clean integration
