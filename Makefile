build:
	@go build -o bin/gfs ./cmd

run: build
	@./bin/gfs

test:
	@go test ./... -v