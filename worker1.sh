#!/bin/sh
 ./worker &
 sleep 5
 echo "START DROP"

 iptables -A INPUT -m statistic --mode random --probability 0.5 -j DROP
#  iptables -A OUTPUT -m statistic --mode random --probability 1 -j DROP

 tail -f /dev/null