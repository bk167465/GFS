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

run-local-master:
	@go run ./cmd/main.go master -port=5000

run-local-cs1:
	@go run ./cmd/main.go chunk -id=cs1 -port=5001 -master=localhost:5000

run-local-cs2:
	@go run ./cmd/main.go chunk -id=cs2 -port=5002 -master=localhost:5000

run-local-cs3:
	@go run ./cmd/main.go chunk -id=cs3 -port=5003 -master=localhost:5000

upload-test-file:
	@go run ./cmd/main.go client -master=localhost:5000 -chunks=cs1=localhost:5001,cs2=localhost:5002,cs3=localhost:5003 -file=testfile.txt
