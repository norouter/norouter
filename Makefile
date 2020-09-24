.DEFAULT_GOAL := binaries

binaries: bin/norouter

bin/norouter:
	CGO_ENABLED=0 go build -o $@ ./cmd/norouter
	LANG=C LC_ALL=C file $@ | grep -qw "statically linked"

clean:
	rm -rf bin

integration:
	./integration/test-internal-agent.sh
	./integration/test-router.sh

.PHONY: bin/norouter clean integration
