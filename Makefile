export IPFS_API=ipfs.io

all:

install:
	go install

ipfs-pack: build

build:
	go build

test: ipfs-pack
	cd test/sharness && make

.PHONY: test
