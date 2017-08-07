.PHONY: build clean data test

build:
	go build -v

clean:
	rm -fr perfdb sample-docs/sample-docs build

data:
	go build -v -o sample-docs/sample-docs ./sample-docs
	./sample-docs/sample-docs

fmt:
	gofmt -w -s *.go

test:
	go test -cover -race -v
