.PHONY: build clean data test

build:
	go build -v

clean:
	rm -fr perfdb build

fmt:
	gofmt -w -s *.go

test:
	go test -cover -race -v
