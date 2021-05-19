build:
		# Get dependencies
		go mod download

		# Make the bin directory
		mkdir -p bin

		# Build for linux and windows
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/spotify_sync_windows.exe
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/spotify_sync_linux
clean:
		go clean
		rm -R bin
