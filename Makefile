.PHONY: build test install clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o try .

test:
	go test -v ./...

install:
	go install $(LDFLAGS) .

clean:
	rm -f try

# Cross-compilation
.PHONY: build-all
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o try-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o try-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o try-linux-amd64 .
