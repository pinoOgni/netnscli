# Define the output directory and binary name
BIN_DIR := bin
BINARY := netnscli

# Define the Go build command
BUILD_CMD := go build -a -o $(BIN_DIR)/$(BINARY)

# Default target to build the binary
all: build

# Target to build the binary
build:
	@mkdir -p $(BIN_DIR)
	$(BUILD_CMD)

# Target to clean the build artifacts
clean:
	@rm -rf $(BIN_DIR)

# Target to run the binary
run: build
	$(BIN_DIR)/$(BINARY)

.PHONY: all build clean run
