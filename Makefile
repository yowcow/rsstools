.PHONY: all test

all:
	make -C httpworker
	make -C rssworker

test:
	go test -v ./httpworker ./rssworker
