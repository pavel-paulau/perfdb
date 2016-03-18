.PHONY: build clean data test

build:
	go build -v

clean:
	rm -f perfdb sample-docs/sample-docs

data:
	go build -v -o sample-docs/sample-docs ./sample-docs
	./sample-docs/sample-docs

test:
	go test -cover -race -v
