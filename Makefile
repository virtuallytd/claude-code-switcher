.PHONY: build install clean test release snapshot

# Build variables
BINARY_NAME=ccs
INSTALL_PATH=/usr/local/bin

build:
	go build -ldflags "-X main.version=dev" -o $(BINARY_NAME)

install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	go clean

test:
	go test ./...

run:
	go run . switch

# Build a snapshot release locally (no git tags required)
snapshot:
	goreleaser release --snapshot --clean

# Full local release (requires git tag)
release:
	goreleaser release --clean

.DEFAULT_GOAL := build
