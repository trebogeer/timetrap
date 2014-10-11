#Build

MAKEFLAGS = -s

.PHONY: all build test clean unit_test 

all: build test

build:
	go install ./...

test: unit_test 

clean:
	go clean -i ./...

unit_test:
	go test ./...


