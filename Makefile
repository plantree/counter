.PHONY: dep build clean run lint test

all: build

# get the dependencies
dep:
	go mod download

build: dep
	go build -o ./build/counter-service

run: build 
	./build/counter-service

test:
	go test -v

test-coverage:
	go test -coverprofile coverage.out 

clean:
	rm -rf build/*
	rm -rf log/*