proto-gen:
	@echo "Generating protobuf files..."
	@protoc -I protos --go_out=protos --go-grpc_out=protos protos/*.proto

proto-clean:
	@echo "Cleaning generated protobuf files..."
	@find protos -name "*.pb.go" -delete

build: proto-gen
	@go build -o bin/gfs ./cmd

run: build
	@./bin/gfs

test: proto-gen
	@go test ./... -v

clean: proto-clean
	@rm -rf bin/