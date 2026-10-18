PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

.PHONY: build init start install uninstall

install: build
	sudo install -Dm755 bin/gitport/main $(BINDIR)/gitport
	sudo setcap 'cap_net_bind_service=+ep' $(BINDIR)/gitport

build:
	go build -o bin/gitport/main cmd/gitport/main.go

init:
	sudo ./bin/gitport/main init

start:
	sudo ./bin/gitport/main start

uninstall:
	sudo rm -f $(BINDIR)/gitport
