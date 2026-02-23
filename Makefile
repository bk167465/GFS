build:
	@go build -o bin/gfs

run: build
	@./bin/gfs

test:
	@go test ./.. -v