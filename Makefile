.PHONY: all test

all:
	godep save -v ./...

test:
	go test -v ./httpworker ./rssworker ./logworker ./ircworker
