.PHONY: all test

all:
	make -C httpworker
	make -C rssworker
	make -C logworker

test:
	go test -v ./httpworker ./rssworker ./logworker
