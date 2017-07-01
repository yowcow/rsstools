.PHONY: all test

all:
	make -C httpworker
	make -C rssworker
	make -C logworker
	make -C ircworker

test:
	go test -v ./httpworker ./rssworker ./logworker ./ircworker
