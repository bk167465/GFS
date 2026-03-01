package main

import (
	"log"
	"os"

	"github.com/bk167465/GFS/cmd/run_grpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: grpc [master|chunk|client] [flags]")
	}

	switch os.Args[1] {
	case "master":
		run_grpc.MasterCommand()
	case "chunk":
		run_grpc.ChunkCommand()
	case "client":
		run_grpc.ClientCommand()
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}
