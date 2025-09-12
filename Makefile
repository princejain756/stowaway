BINARY := stowaway
GO      ?= go
GOFLAGS ?=

.PHONY: all build test vet fmt install clean

all: build

build:
	$(GO) build $(GOFLAGS) -o $(BINARY) .

test:
	$(GO) test $(GOFLAGS) ./...

vet:
	$(GO) vet $(GOFLAGS) ./...

fmt:
	$(GO) fmt ./...

install:
	$(GO) install $(GOFLAGS) .

clean:
	rm -f $(BINARY)
	$(GO) clean -testcache
