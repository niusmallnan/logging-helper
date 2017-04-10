#!/bin/sh

set -e

for i in $(curl -s --unix /var/run/docker.sock http://localhost/info | jq -r .DockerRootDir) /var/lib/docker /run /var/run; do
    for m in $(tac /proc/mounts | awk '{print $2}' | grep ^${i}/); do
        if [ "$m" != "/var/run/docker.sock"  ]; then
            umount $m || true
        fi
    done
done

exec "$@"
