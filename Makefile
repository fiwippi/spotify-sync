 # Go parameters
BINARY_NAME=spotify_sync

build:
		# Get dependencies
		go mod download

		# Make the bin directory
		mkdir -p bin

		# Build for linux and windows
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)_server_windows.exe cmd/server.go
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)_client_windows.exe cmd/client.go
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)_view_windows.exe cmd/view.go

		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)_server_linux cmd/server.go
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)_client_linux cmd/client.go
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)_view_linux cmd/view.go
clean:
		go clean
		rm -R bin
