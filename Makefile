build:
	GOOS=linux GOARCH=arm64 go build -ldflags '-s -w' -o bin/main ./cmd