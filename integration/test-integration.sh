#!/bin/bash
# TODO: rewrite from scratch, either in bats or in go
set -eu -o pipefail

cd "$(dirname $0)/.."

LABEL="norouter-test-integration"

pid=""
cleanup() {
#  if [[ -n "$pid" ]]; then
#    echo "trap"
#    sleep infinity
#  fi
	echo "Cleaning up..."
	set +e
	if [[ -n "$pid" && -d "/proc/$pid" ]]; then kill $pid; fi
	docker rm -f $(docker ps -f label=$LABEL -a -q) || true
	make clean
	sleep 3
	if [[ -n "$pid" && -d "/proc/$pid" ]]; then echo "process still running?"; exit 1; fi
	set -e
}
cleanup
trap "cleanup" EXIT

make
docker run -l $LABEL -d --name host1 -v "$(pwd)/bin:/mnt:ro" nginx:1.19.2-alpine
docker run -l $LABEL -d --name host2 -v "$(pwd)/bin:/mnt:ro" httpd:2.4.46-alpine
docker run -l $LABEL -d --name host3 -v "$(pwd)/bin:/mnt:ro" caddy:2.1.1-alpine

docker exec host1 apk add --no-cache iperf3
docker exec host2 apk add --no-cache iperf3

# dind to emulate remote network 192.168.95.0/24, which isn't accessible from host
docker run -l $LABEL -d --name dind1 -v "$(pwd)/bin:/mnt:ro" --privileged -v dind1-vol:/var/lib/docker docker:19.03.13-dind
sleep 10; until docker exec dind1 docker info; do sleep 10; done
docker exec dind1 docker network create dind1-subnet95 --subnet=192.168.95.0/24
docker exec -t dind1 docker run --network dind1-subnet95 -d --name dind1-bastion -v "/mnt:/mnt:ro" alpine sleep infinity
docker exec -t dind1 docker run --network dind1-subnet95 -d --name dind1-wordpress --hostname dind1-wordpress --ip=192.168.95.101 wordpress:5.5.3
docker exec -t dind1 docker run --network dind1-subnet95 -d --name dind1-mediawiki --hostname dind1-mediawiki --ip=192.168.95.102 mediawiki:1.35.0

docker run -l $LABEL -d --name dind2 -v "$(pwd)/bin:/mnt:ro" --privileged -v dind2-vol:/var/lib/docker docker:19.03.13-dind
sleep 10; until docker exec dind2 docker info; do sleep 10; done
docker exec dind2 docker network create dind2-subnet96 --subnet=192.168.96.0/24
docker exec -t dind2 docker run --network dind2-subnet96 -d --name dind2-bastion -v "/mnt:/mnt:ro" alpine sleep infinity
docker exec -t dind2 docker run --network dind2-subnet96 -d --name dind2-wordpress --hostname dind2-wordpress --ip=192.168.96.101 wordpress:5.5.3

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


targets=(http://host1:8080 http://host2:8080 http://host3:8080 http://dind1-wordpress.dind1-subnet95 http://192.168.95.102 http://dind2-wordpress.dind2-subnet96)
echo "Testing http proxy mode"
set -x
for ((i = 0; i < $N; i++)); do
  for f in ${targets[@]}; do
    curl -fsSL -o /dev/null --proxy http://127.0.0.1:18080 ${f}
  done
done
set +x

echo "Testing http proxy mode (HTTP TUNNEL)"
set -x
for ((i = 0; i < $N; i++)); do
  for f in ${targets[@]}; do
    curl -fsS -o /dev/null --proxy http://127.0.0.1:18080 --proxytunnel ${f}
  done
done
set +x

echo "Testing SOCKS4a mode"
set -x
for ((i = 0; i < $N; i++)); do
  for f in ${targets[@]}; do
    curl -fsS -o /dev/null --socks4a http://127.0.0.1:18081 ${f}
  done
done
set +x

echo "Testing SOCKS5h mode"
set -x
for ((i = 0; i < $N; i++)); do
  for f in ${targets[@]}; do
    curl -fsS -o /dev/null --socks5-hostname http://127.0.0.1:18081 ${f}
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
