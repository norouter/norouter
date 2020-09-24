#!/bin/bash
set -eu -o pipefail

cd "$(dirname $0)/.."

pid=""
cleanup() {
	echo "Cleaning up..."
	set +e
	if [[ -n "$pid" && -d "/proc/$pid" ]]; then kill $pid; fi
	docker rm -f host1 host2 host3
	make clean
	set -e
}
cleanup
trap "cleanup" EXIT

make
docker run -d --name host1 -v "$(pwd)/bin:/mnt:ro" nginx:1.19.2-alpine
docker run -d --name host2 -v "$(pwd)/bin:/mnt:ro" httpd:2.4.46-alpine
docker run -d --name host3 -v "$(pwd)/bin:/mnt:ro" caddy:2.1.1-alpine

: ${DEBUG=}
flags=""
if [[ -n "$DEBUG" ]]; then
	flags="--debug"
fi

./bin/norouter ${flags} ./integration/test-router.yaml &
pid=$!

sleep 3

: ${N=3}
succeeds=0
fails=0

test_wget() {
	for ((i = 0; i < $N; i++)); do
		if wget -q -O- $1 | grep -q "$2"; then
			succeeds=$((succeeds + 1))
		else
			fails=$((fails + 1))
		fi
    for ((j = 1; j <= 3; j++)); do
      if docker exec host${j} wget -q -O- $1 | grep -q "$2"; then
        succeeds=$((succeeds + 1))
      else
        fails=$((fails + 1))
      fi
    done
  done
}

# Connect to host1 (nginx)
test_wget http://127.0.42.101:8080 "Welcome to nginx"
# Connect to host2 (Apache httpd)
test_wget http://127.0.42.102:8080 "It works"
# Connect to host3 (Caddy) 
test_wget http://127.0.42.103:8080 "Caddy"

echo "tests: $((N * 4 * 3)), succceeds: ${succeeds}, fails: ${fails}"
exit ${fails}
