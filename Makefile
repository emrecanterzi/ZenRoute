BINARY_NAME=zenroute
BIN_DIR=bin

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/zenroute

run: build
	./$(BIN_DIR)/$(BINARY_NAME)
