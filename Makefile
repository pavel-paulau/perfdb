.PHONY: build clean data test

build:
	go build -race -v

clean:
	rm -f perfdb sample-docs/sample-docs

data:
	go build -race -v -o sample-docs/sample-docs ./sample-docs
	./sample-docs/sample-docs

test:
	go test -cover -v
