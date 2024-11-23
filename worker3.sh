#!/bin/sh
# sleep 2
# ./worker &

# sleep 15
# echo WORKER PAUSED
# pkill -19 worker # STP
# sleep 15
# echo WORKER RESUMED
# pkill -18 worker # CONT
# tail -f /dev/null