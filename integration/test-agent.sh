#!/bin/bash
# A script for testing `norouter internal agent` without the main process
#
set -eu -o pipefail

cd "$(dirname $0)/.."

pid=""
cleanup() {
	echo "Cleaning up..."
	set +e
	if [[ -n "$pid" && -d "/proc/$pid" ]]; then kill $pid; fi
	docker rm -f host1 host2
	make clean
	set -e
}
cleanup
trap "cleanup" EXIT

make
docker run -d --name host1 -v "$(pwd)/bin:/mnt:ro" nginx:1.19.2-alpine
docker run -d --name host2 -v "$(pwd)/bin:/mnt:ro" httpd:2.4.46-alpine

: ${DEBUG=}
flags=""
if [[ -n "$DEBUG" ]]; then
	flags="--debug"
fi

config1='{"me":"127.0.42.101","others":[{"ip":"127.0.42.102","port":8080,"proto":"tcp"}],"forwards":[{"listen_port":8080,"connect_ip":"127.0.0.1","connect_port":80,"proto":"tcp"}]}'
config2='{"me":"127.0.42.102","others":[{"ip":"127.0.42.101","port":8080,"proto":"tcp"}],"forwards":[{"listen_port":8080,"connect_ip":"127.0.0.1","connect_port":80,"proto":"tcp"}]}'

dpipe \
	docker exec -i host1 /mnt/norouter ${flags} agent --automated \
	--debug-init-config "${config1}" \
	= \
	docker exec -i host2 /mnt/norouter ${flags} agent --automated \
	--debug-init-config "${config2}" &
pid=$!

sleep 2

: ${N=10}
succeeds=0
fails=0
# Connect to host1 (127.0.42.101, nginx) from host2
for ((i = 0; i < $N; i++)); do
	if docker exec host2 wget -q -O- http://127.0.42.101:8080 | grep -q "Welcome to nginx"; then
		succeeds=$((succeeds + 1))
	else
		fails=$((fails + 1))
	fi
done

# Connect to host2 (127.0.42.102, Apache httpd) from host1
for ((i = 0; i < $N; i++)); do
	if docker exec host1 wget -q -O- http://127.0.42.102:8080 | grep -q "It works"; then
		succeeds=$((succeeds + 1))
	else
		fails=$((fails + 1))
	fi
done

echo "tests: $((N * 2)), succceeds: ${succeeds}, fails: ${fails}"
exit ${fails}
