.PHONY:
	build clean test

build:
	go build -race -v

clean:
	rm -f perfkeeper

test:
	go test -cover -v
