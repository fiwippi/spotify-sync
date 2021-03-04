 # Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOGENERATE=$(GOCMD) generate
BINARY_NAME=spotify_sync

build:
		# Get dependencies
		$(GOMOD) download

		# Make the bin directory
		mkdir -p bin

		# Build for linux and windows
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_server_windows.exe cmd/server.go
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_client_windows.exe cmd/client.go
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_view_windows.exe cmd/view.go

		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_server_linux cmd/server.go
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_client_linux cmd/client.go
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)_view_linux cmd/view.go
clean:
		$(GOCLEAN)
		rm -R bin
