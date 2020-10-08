.PHONY: build
build:
	go build -i -v -o main 

test:
	go test -race -coverprofile=coverage.txt