.PHONY: build install clean test

# Build variables
BINARY_NAME=ccs
INSTALL_PATH=/usr/local/bin

build:
	go build -o $(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test ./...

run:
	go run . switch

.DEFAULT_GOAL := build
