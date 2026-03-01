package run_grpc

import (
	"flag"
	"log"
	"os"

	"github.com/bk167465/GFS/internal/chunkserver"
)

// runChunkServer starts a chunk server
func runChunkServer(serverID string, port string) error {
	cs := chunkserver.NewChunkServer(serverID)
	server := chunkserver.NewChunkServerGRPC(cs)
	return server.StartServer(serverID, port)
}

// ChunkCommand handles the chunk server startup
func ChunkCommand() {
	chunkCmd := flag.NewFlagSet("chunk", flag.ExitOnError)
	serverID := chunkCmd.String("id", "cs1", "Chunk server ID")
	port := chunkCmd.String("port", "5001", "Port for chunk server")

	// use os.Args directly: skip program name and subcommand
	chunkCmd.Parse(os.Args[2:])

	log.Printf("Starting Chunk Server: %s...", *serverID)
	if err := runChunkServer(*serverID, *port); err != nil {
		log.Fatalf("Chunk server error: %v", err)
	}
}
