.PHONY: all test

all:
	godep save -v ./...

test:
	go test -v ./httpworker ./rssworker ./itemworker ./broadcaster
