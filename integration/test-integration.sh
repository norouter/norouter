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

docker exec host1 apk add --no-cache iperf3
docker exec host2 apk add --no-cache iperf3

: ${DEBUG=}
flags=""
if [[ -n "$DEBUG" ]]; then
	flags="--debug"
fi

./bin/norouter ${flags} ./integration/test-integration.yaml &
pid=$!

sleep 5

: ${N=3}
succeeds=0
fails=0

test_wget() {
	for ((i = 0; i < $N; i++)); do
		tmp=$(mktemp)
		wget -q -O- $1 >$tmp
		if grep -q "$2" $tmp; then
			succeeds=$((succeeds + 1))
		else
			fails=$((fails + 1))
		fi
		rm -f $tmp
		for ((j = 1; j <= 3; j++)); do
			tmp=$(mktemp)
			docker exec host${j} wget -q -O- $1 >$tmp
			if grep -q "$2" $tmp; then
				succeeds=$((succeeds + 1))
			else
				fails=$((fails + 1))
				echo "<Unexpected output (lacks \"$2\")>"
				cat $tmp
				echo "</Unexpected output (lacks \"$2\")>"
			fi
			rm -f $tmp
		done
	done
}

echo "Testing loopback mode"
# Connect to host1 (nginx)
test_wget http://127.0.42.101:8080 "Welcome to nginx"
# Connect to host2 (Apache httpd)
test_wget http://127.0.42.102:8080 "It works"
# Connect to host3 (Caddy)
test_wget http://127.0.42.103:8080 "Caddy"

echo "tests: $((N * 4 * 3)), succceeds: ${succeeds}, fails: ${fails}"
if [ ${fails} -ne "0" ]; then
  exit ${fails}
fi

echo "Testing http proxy mode"
set -x
for ((i = 0; i < $N; i++)); do
  for f in host1 host2 host3; do
    curl -fsS -o /dev/null --proxy http://127.0.0.1:18080 http://${f}:8080
  done
done
set +x

echo "Testing http proxy mode (HTTP TUNNEL)"
set -x
for ((i = 0; i < $N; i++)); do
  for f in host1 host2 host3; do
    curl -fsS -o /dev/null --proxy http://127.0.0.1:18080 --proxytunnel http://${f}:8080
  done
done
set +x

echo "iperf3 from host2 to host1"
docker exec host1 iperf3 -s > /dev/null &
iperf3_exec_pid=$!
sleep 2
docker exec host2 iperf3 -p 15201 -c 127.0.42.101
set +e
kill ${iperf3_exec_pid}
set -e
exit 0
