# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=rpi-heating-controller
OUTPUT_DIR=bin
APP_DIR=app


# Target host to transfer binary (use `make transfer HOST=<hostname>` to specify)
HOST=

# Build for Raspberry Pi (ARM)
build:
	env GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(OUTPUT_DIR)

# Run the application (optional, modify as needed)
run:
	go run .

# Transfer binary to the specified host using `scp`
transfer:
	scp $(OUTPUT_DIR)/$(BINARY_NAME) $(HOST):

# Default target when running `make` without any arguments
default: build