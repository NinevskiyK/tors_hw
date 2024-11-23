#!/bin/sh
./worker &

sleep 1
echo "START DROP"

# tc qdisc add dev eth0 root netem loss random 10%
# tc qdisc add dev eth0 root netem delay 1000ms
# tc qdisc add dev eth0 root netem duplicate 10%
tc qdisc add dev eth0 root netem delay 1000ms reorder 25% 50%


tail -f /dev/null