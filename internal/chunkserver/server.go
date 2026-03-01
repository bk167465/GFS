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

func (csg *ChunkServerGRPC) PushData(
	ctx context.Context,
	req *pb.PushDataRequest,
) (*pb.PushDataResponse, error) {

	csg.chunkServer.mu.Lock()
	defer csg.chunkServer.mu.Unlock()

	// Store in temporary buffer (NOT disk yet)
	csg.chunkServer.writeBuffer[req.DataId] = req.Data

	return &pb.PushDataResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (csg *ChunkServerGRPC) CommitWrite(
	ctx context.Context,
	req *pb.CommitWriteRequest,
) (*pb.CommitWriteResponse, error) {
	err := csg.chunkServer.ApplyWrite(req.Handle, req.DataId, req.Offset)
	if err != nil {
		return &pb.CommitWriteResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// If already forwarded, stop here
	if req.FromPrimary {
		return &pb.CommitWriteResponse{
			Success: true,
			Error:   "",
		}, nil
	}

	// Forward to secondaries
	for _, addr := range req.Secondaries {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return &pb.CommitWriteResponse{
				Success: false,
				Error:   fmt.Sprintf("connect secondary %s: %v", addr, err),
			}, nil
		}

		client := pb.NewChunkServiceClient(conn)

		_, err = client.CommitWrite(ctx, &pb.CommitWriteRequest{
			Handle:      req.Handle,
			DataId:      req.DataId,
			Offset:      req.Offset,
			FromPrimary: true,
		})

		conn.Close()

		if err != nil {
			return &pb.CommitWriteResponse{
				Success: false,
				Error:   fmt.Sprintf("commit failed on %s: %v", addr, err),
			}, nil
		}
	}

	return &pb.CommitWriteResponse{
		Success: true,
		Error:   "",
	}, nil
}
