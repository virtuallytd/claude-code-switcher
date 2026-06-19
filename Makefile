VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build clean

build:
	go build -ldflags '$(LDFLAGS)' -o ccs .

clean:
	rm -f ccs
