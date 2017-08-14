.PHONY: all test

all:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

test:
	go test ./httpworker ./rssworker ./itemworker ./broadcaster
