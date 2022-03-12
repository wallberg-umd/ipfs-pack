export IPFS_API=ipfs.io

all:

install:
	go install

build:
	go build

test:
	cd test/sharness && make

.PHONY: test
