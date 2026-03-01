package run_grpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bk167465/GFS/internal/client"
)

// ClientCommand handles the gRPC client file operations
func ClientCommand() {
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	masterAddr := clientCmd.String("master", "localhost:5000", "Master server address")
	chunksStr := clientCmd.String("chunks", "cs1=localhost:5001,cs2=localhost:5002,cs3=localhost:5003", "Chunk servers as comma-separated id=address pairs")
	filename := clientCmd.String("file", "", "File to upload")

	// use os.Args directly: skip program name and subcommand
	clientCmd.Parse(os.Args[2:])

	// allow positional filename (in case flag parsing removed it)
	if *filename == "" && clientCmd.NArg() > 0 {
		*filename = clientCmd.Arg(0)
	}

	if *filename == "" {
		log.Fatalf("Please specify a file with -file flag or as positional argument")
	}

	// if file not found, try adding .txt extension as a convenience
	if _, err := os.Stat(*filename); os.IsNotExist(err) {
		try := *filename + ".txt"
		if _, err2 := os.Stat(try); err2 == nil {
			*filename = try
		}
	}

	// Parse chunk server addresses
	chunkServerAddrs := make(map[string]string)
	for _, pair := range strings.Split(*chunksStr, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			log.Fatalf("Invalid chunk server format: %s", pair)
		}
		chunkServerAddrs[parts[0]] = parts[1]
	}

	// Create gRPC client
	gClient, err := client.NewGRPCClient(*masterAddr, chunkServerAddrs)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer gClient.Close()

	// Upload file
	log.Printf("Uploading file: %s\n", *filename)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = uploadViaGRPC(ctx, gClient, *filename)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Verify file metadata
	fileMeta, err := gClient.GetFileMetadata(ctx, *filename)
	if err != nil {
		log.Fatalf("Failed to get file metadata: %v", err)
	}

	fmt.Println("\n=== Upload Complete ===")
	fmt.Printf("File: %s\n", fileMeta.Filename)
	for i, chunk := range fileMeta.Chunks {
		fmt.Printf("  Chunk %d: %s\n", i, chunk.Handle)
		fmt.Printf("    Locations: %v\n", chunk.Locations)
	}
}

// uploadViaGRPC uploads a file using gRPC
func uploadViaGRPC(ctx context.Context, gClient *client.GRPCClient, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	const chunkSize = 1024 * 1024 // 1MB
	var offset int64 = 0

	for {
		buffer := make([]byte, chunkSize)
		n, err := file.Read(buffer)

		if n == 0 {
			break
		}

		// Allocate chunk via Master
		chunkMeta, err := gClient.AllocateChunk(ctx, filename)
		if err != nil {
			return fmt.Errorf("AllocateChunk failed: %w", err)
		}

		fmt.Printf("Allocated chunk %s\n", chunkMeta.Handle)
		fmt.Printf("  Primary: %s\n", chunkMeta.Primary)
		fmt.Printf("  Replicas: %v\n", chunkMeta.Locations)

		// Write chunk using new Push + Commit flow
		err = gClient.WriteChunk(ctx, filename, buffer[:n], offset)
		if err != nil {
			return fmt.Errorf("WriteChunk failed: %w", err)
		}

		fmt.Printf("Chunk %s written successfully\n", chunkMeta.Handle)

		offset += int64(n)

		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("file read error: %w", err)
		}

		if err != nil { // EOF
			break
		}
	}

	return nil
}
