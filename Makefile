.PHONY: all test

all:
	go get github.com/golang/dep/cmd/dep
	dep ensure -v

test:
	go test ./httpworker ./rssworker ./itemworker ./broadcaster
