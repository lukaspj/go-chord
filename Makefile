# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
PROTOC=protoc -I $(GOPATH)/src -I api/
BINARY_DIR=bin/
BINARY_WINDOWS=$(BINARY_DIR)chord.exe
BINARY_UNIX=$(BINARY_DIR)$(BINARY_NAME)_unix
PKG=github.com/lukaspj/go-chord

all: test build

build:
	$(GOBUILD) -o $(BINARY_WINDOWS) -v $(PKG)/cmd/chord
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_WINDOWS) -v ./...
	./$(BINARY_WINDOWS)
deps:
	$(GOGET) github.com/lukaspj/go-logging/logging
generate:
	$(PROTOC) api/chord.proto --go_out=plugins=grpc:api

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v