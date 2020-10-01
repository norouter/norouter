#!/bin/bash
# Shows the max throughput of stdio across two `docker exec` instances,
# without running NoRouter.
#
set -eu -o pipefail

pid=""
cleanup() {
	echo "Cleaning up..."
	set +e
	docker rm -f bench-c0 bench-c1
	set -e
}
cleanup
trap "cleanup" EXIT

docker run -d --name bench-c0 alpine sleep infinity
docker run -d --name bench-c1 alpine sleep infinity

# Transfer 1GiB
begin=$(date +%s)
mega=$((1024 * 1024))
count=1024
docker exec bench-c0 dd if=/dev/zero of=/dev/stdout bs=${mega} count=${count} status=none | docker exec -i bench-c1 dd if=/dev/stdin of=/dev/null bs=${mega} status=none
end=$(date +%s)


# FIXME: floating point math
total=$(($end - $begin))
echo "Transferred ${count} MiBs in ${total} seconds"

mibps=$(($count / $total))
echo "Throughput: ${mibps} MiBps"
