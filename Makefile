.PHONY: all test

all:
	which dep || go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

test:
	go test ./...
