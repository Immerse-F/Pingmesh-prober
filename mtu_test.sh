#!/bin/bash

IP="$1"
MIN=1000
MAX=65535

if [ -z "$IP" ]; then
  echo "Usage: ./mtu_test.sh <IP>"
  exit 1
fi

while [ $MIN -lt $MAX ]; do
  MID=$(((MIN+MAX+1)/2))
  ping_output=$(ping -c 1 -M do -s $MID -W 0.8 $IP 2>&1)
if [[ $ping_output == *"1 packets transmitted, 1 received"* ]]; then
  MIN=$MID
else
  MAX=$((MID-1))
fi
done

echo "MTU for $IP is $(($MIN + 48))"
