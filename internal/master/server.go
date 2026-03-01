package master

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/bk167465/GFS/internal/common"
	pb "github.com/bk167465/GFS/protos/pb"
	"google.golang.org/grpc"
)

// MasterServer implements the gRPC MasterService
type MasterServer struct {
	pb.UnimplementedMasterServiceServer
	master *Master
}

// NewMasterServer creates a new gRPC Master server
func NewMasterServer(m *Master) *MasterServer {
	return &MasterServer{
		master: m,
	}
}

// AllocateChunk allocates a new chunk for a file
func (ms *MasterServer) AllocateChunk(ctx context.Context, req *pb.AllocateChunkRequest) (*pb.AllocateChunkResponse, error) {
	chunk := ms.master.AllocateChunk(req.Filename)

	pbChunk := &pb.ChunkMetadata{
		Handle:    string(chunk.Handle),
		Locations: locationsToStrings(chunk.Locations),
		Version:   int32(chunk.Version),
	}

	return &pb.AllocateChunkResponse{
		Chunk: pbChunk,
		Error: "",
	}, nil
}

// GetChunkLocations returns the servers storing a chunk
func (ms *MasterServer) GetChunkLocations(ctx context.Context, req *pb.GetChunkLocationsRequest) (*pb.ChunkLocationResponse, error) {
	meta, ok := ms.master.GetFileMetadata(req.Filename)
	if !ok {
		return &pb.ChunkLocationResponse{
			Chunk: nil,
			Error: "file not found",
		}, nil
	}

	if len(meta.Chunks) == 0 {
		return &pb.ChunkLocationResponse{
			Chunk: nil,
			Error: "no chunks found for file",
		}, nil
	}

	firstChunk := meta.Chunks[0]
	pbChunk := &pb.ChunkMetadata{
		Handle:    string(firstChunk.Handle),
		Locations: locationsToStrings(firstChunk.Locations),
		Version:   int32(firstChunk.Version),
	}

	return &pb.ChunkLocationResponse{
		Chunk: pbChunk,
		Error: "",
	}, nil
}

// GetFileMetadata returns all chunks for a file
func (ms *MasterServer) GetFileMetadata(ctx context.Context, req *pb.GetFileMetadataRequest) (*pb.GetFileMetadataResponse, error) {
	meta, ok := ms.master.GetFileMetadata(req.Filename)
	if !ok {
		return &pb.GetFileMetadataResponse{
			Metadata: nil,
			Error:    "file not found",
		}, nil
	}

	pbChunks := make([]*pb.ChunkMetadata, len(meta.Chunks))
	for i, chunk := range meta.Chunks {
		pbChunks[i] = &pb.ChunkMetadata{
			Handle:    string(chunk.Handle),
			Locations: locationsToStrings(chunk.Locations),
			Version:   int32(chunk.Version),
		}
	}

	pbMeta := &pb.FileMetadata{
		Filename: meta.Filename,
		Chunks:   pbChunks,
	}

	return &pb.GetFileMetadataResponse{
		Metadata: pbMeta,
		Error:    "",
	}, nil
}

// StartServer starts the gRPC Master server
func (ms *MasterServer) StartServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMasterServiceServer(grpcServer, ms)

	log.Printf("Master gRPC server listening on port %s\n", port)

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Helper function to convert ServerID slice to strings
func locationsToStrings(locations []common.ServerID) []string {
	result := make([]string, len(locations))
	for i, loc := range locations {
		result[i] = string(loc)
	}
	return result
}
