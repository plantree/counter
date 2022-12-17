.PHONY: build clean run lint test

all: build

build:
	cd cmd && go build -o ../build/counter-service

run: build 
	./build/counter-service

clean:
	rm -rf build/*
	rm -rf log/*