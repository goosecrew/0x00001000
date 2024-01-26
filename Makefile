build:
	GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o bin/rs232-reader ./cmd
