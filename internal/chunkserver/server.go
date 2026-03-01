package chunkserver

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/bk167465/GFS/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ChunkServerGRPC implements the gRPC ChunkService
type ChunkServerGRPC struct {
	pb.UnimplementedChunkServiceServer
	chunkServer *ChunkServer
}

// NewChunkServerGRPC creates a new gRPC Chunk server
func NewChunkServerGRPC(cs *ChunkServer) *ChunkServerGRPC {
	return &ChunkServerGRPC{
		chunkServer: cs,
	}
}

// WriteChunk writes data to a chunk and optionally replicates to secondaries
func (csg *ChunkServerGRPC) WriteChunk(ctx context.Context, req *pb.WriteChunkRequest) (*pb.WriteChunkResponse, error) {
	// Write locally
	if err := csg.chunkServer.WriteChunk(req.Handle, req.Data); err != nil {
		return &pb.WriteChunkResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to write chunk: %v", err),
		}, nil
	}

	// If there are replicas, replicate to them
	if len(req.Replicas) > 0 {
		if err := csg.replicateToServers(ctx, req.Handle, req.Data, req.Replicas); err != nil {
			return &pb.WriteChunkResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to replicate: %v", err),
			}, nil
		}
	}

	return &pb.WriteChunkResponse{
		Success: true,
		Error:   "",
	}, nil
}

// ReadChunk retrieves data from a chunk
func (csg *ChunkServerGRPC) ReadChunk(ctx context.Context, req *pb.ReadChunkRequest) (*pb.ReadChunkResponse, error) {
	data, err := csg.chunkServer.ReadChunk(req.Handle)
	if err != nil {
		return &pb.ReadChunkResponse{
			Data:  nil,
			Error: fmt.Sprintf("failed to read chunk: %v", err),
		}, nil
	}

	return &pb.ReadChunkResponse{
		Data:  data,
		Error: "",
	}, nil
}

// CheckChunk verifies if server has a chunk
func (csg *ChunkServerGRPC) CheckChunk(ctx context.Context, req *pb.CheckChunkRequest) (*pb.CheckChunkResponse, error) {
	exists := csg.chunkServer.HasChunk(req.Handle)
	return &pb.CheckChunkResponse{
		Exists: exists,
	}, nil
}

// StartServer starts the gRPC Chunk server
func (csg *ChunkServerGRPC) StartServer(serverID string, port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChunkServiceServer(grpcServer, csg)

	log.Printf("ChunkServer %s gRPC server listening on port %s\n", serverID, port)

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// replicateToServers replicates data to secondary servers
func (csg *ChunkServerGRPC) replicateToServers(ctx context.Context, handle string, data []byte, secondaryAddrs []string) error {
	for _, addr := range secondaryAddrs {
		// Connect to secondary
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("failed to connect to secondary %s: %w", addr, err)
		}
		defer conn.Close()

		client := pb.NewChunkServiceClient(conn)

		// Send write request to secondary (no further replication)
		_, err = client.WriteChunk(ctx, &pb.WriteChunkRequest{
			Handle:   handle,
			Data:     data,
			Replicas: []string{}, // Don't replicate further down the chain
		})
		if err != nil {
			return fmt.Errorf("failed to replicate to %s: %w", addr, err)
		}
	}
	return nil
}
