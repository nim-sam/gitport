.PHONY: build init start

build:
	go build -o bin/gitport/main cmd/gitport/main.go

init:
	sudo ./bin/gitport/main init

start:
	sudo ./bin/gitport/main start
