APP_NAME := gitport
BIN_DIR  := bin
CMD_DIR  := ./cmd

GO := go

.PHONY: all build install clean run

all: build

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

install: build
	install -m 755 $(BIN_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)

run: build
	./$(BIN_DIR)/$(APP_NAME)

clean:
	rm -rf $(BIN_DIR)
