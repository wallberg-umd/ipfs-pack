#!/bin/bash

rm -rf packtest
mkdir -p packtest
cd packtest
random-files -seed=42 -files=6 -depth=3 -dirs=5 stuff
ipfs-pack make

ipfs-pack serve --json > serve_out &
echo $! > serve-pid
sleep 1

PEERID=$(cat serve_out | awk 'match($0, /"PeerID":"[^"]+/){print substr($0, RSTART+10, RLENGTH-10)}')
ADDR=$(cat serve_out | awk 'match($0, /"Addresses":\["[^"]+/){print substr($0, RSTART+14, RLENGTH-14)}')
echo peerid $PEERID
echo addr $ADDR

rm -rf ../ipfs
export IPFS_PATH=../ipfs
ipfs init
ipfs bootstrap rm --all
ipfs config --json Discovery.MDNS.Enabled false
ipfs daemon &
echo $! > ipfs-pid
sleep 1
ipfs swarm connect $ADDR

HASH=$(tail -n1 PackManifest | awk '{ print $1 }')

ipfs refs --timeout=20s -r $HASH
if (test $? = 0); then
   echo ipfs refs succeeded
fi

kill $(cat serve-pid)
kill $(cat ipfs-pid)
