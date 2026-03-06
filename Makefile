BINARY_NAME=zenroute
BIN_DIR=bin

build:
	mkdir -p $(BIN_DIR)
	cp .env $(BIN_DIR)/.env
	cp bypass-domains.txt $(BIN_DIR)/bypass-domains.txt
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/zenroute

run: build
	./$(BIN_DIR)/$(BINARY_NAME)

build-windows:
	mkdir -p $(BIN_DIR)
	cp .env $(BIN_DIR)/.env
	cp bypass-domains.txt $(BIN_DIR)/bypass-domains.txt
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-windows.exe ./cmd/zenroute

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-mac ./cmd/zenroute

build-mac-arm:
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-mac-arm64 ./cmd/zenroute

build-all: build-windows build-mac build-mac-arm

clean:
	rm -rf $(BIN_DIR)