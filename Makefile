# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=rpi-heating-controller
OUTPUT_DIR=bin
APP_DIR=app
PROTO_DIR=lib/protobuf
# Specify the output directory for the generated protobuf Go code
OUT_DIR := lib/protobuf/output

# Target host to transfer binary (use `make transfer HOST=<hostname>` to specify)
HOST=

# Get a list of all .proto files in the directory
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# Build for Raspberry Pi (ARM)
build: generate
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

# Generate Go code from all .proto files
generate:
	@mkdir -p $(OUT_DIR)
	@for file in $(PROTO_FILES); do \
		protoc --go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) $$file; \
	done


# Default target when running `make` without any arguments
default: build