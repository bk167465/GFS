package client

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/bk167465/GFS/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient wraps gRPC connections to Master and ChunkServers
type GRPCClient struct {
	masterConn   *grpc.ClientConn
	masterClient pb.MasterServiceClient

	chunkConnections map[string]*grpc.ClientConn
	chunkClients     map[string]pb.ChunkServiceClient
}

// NewGRPCClient initializes gRPC connections
func NewGRPCClient(masterAddr string, chunkServerAddrs map[string]string) (*GRPCClient, error) {
	gc := &GRPCClient{
		chunkConnections: make(map[string]*grpc.ClientConn),
		chunkClients:     make(map[string]pb.ChunkServiceClient),
	}

	// Connect to Master
	masterConn, err := grpc.NewClient(masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master at %s: %w", masterAddr, err)
	}
	gc.masterConn = masterConn
	gc.masterClient = pb.NewMasterServiceClient(masterConn)

	// Connect to ChunkServers
	for serverID, serverAddr := range chunkServerAddrs {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			gc.Close()
			return nil, fmt.Errorf("failed to connect to chunkserver %s at %s: %w", serverID, serverAddr, err)
		}
		gc.chunkConnections[serverID] = conn
		gc.chunkClients[serverID] = pb.NewChunkServiceClient(conn)
	}

	fmt.Printf("GRPCClient connected to Master at %s and %d ChunkServers\n", masterAddr, len(chunkServerAddrs))
	return gc, nil
}

// AllocateChunk requests a new chunk allocation from the Master
func (gc *GRPCClient) AllocateChunk(ctx context.Context, filename string) (*pb.ChunkMetadata, error) {
	if gc.masterClient == nil {
		return nil, fmt.Errorf("not connected to master")
	}

	resp, err := gc.masterClient.AllocateChunk(ctx, &pb.AllocateChunkRequest{
		Filename: filename,
	})
	if err != nil {
		return nil, fmt.Errorf("AllocateChunk RPC failed: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("master error: %s", resp.Error)
	}

	return resp.Chunk, nil
}

// GetChunkLocations retrieves chunk locations from the Master
func (gc *GRPCClient) GetChunkLocations(ctx context.Context, filename string) (*pb.ChunkMetadata, error) {
	if gc.masterClient == nil {
		return nil, fmt.Errorf("not connected to master")
	}

	resp, err := gc.masterClient.GetChunkLocations(ctx, &pb.GetChunkLocationsRequest{
		Filename: filename,
	})
	if err != nil {
		return nil, fmt.Errorf("GetChunkLocations RPC failed: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("master error: %s", resp.Error)
	}

	return resp.Chunk, nil
}

// GetFileMetadata retrieves file metadata from the Master
func (gc *GRPCClient) GetFileMetadata(ctx context.Context, filename string) (*pb.FileMetadata, error) {
	if gc.masterClient == nil {
		return nil, fmt.Errorf("not connected to master")
	}

	resp, err := gc.masterClient.GetFileMetadata(ctx, &pb.GetFileMetadataRequest{
		Filename: filename,
	})
	if err != nil {
		return nil, fmt.Errorf("GetFileMetadata RPC failed: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("master error: %s", resp.Error)
	}

	return resp.Metadata, nil
}

// WriteChunkToServer writes data to a specific chunkserver via gRPC
func (gc *GRPCClient) WriteChunkToServer(ctx context.Context, serverID string, handle string, data []byte) error {
	client, ok := gc.chunkClients[serverID]
	if !ok {
		return fmt.Errorf("chunkserver %s not connected", serverID)
	}

	resp, err := client.WriteChunk(ctx, &pb.WriteChunkRequest{
		Handle:   handle,
		Data:     data,
		Replicas: []string{},
	})
	if err != nil {
		return fmt.Errorf("WriteChunk RPC failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("write failed: %s", resp.Error)
	}

	return nil
}

// ReadChunkFromServer reads data from a specific chunkserver via gRPC
func (gc *GRPCClient) ReadChunkFromServer(ctx context.Context, serverID string, handle string) ([]byte, error) {
	client, ok := gc.chunkClients[serverID]
	if !ok {
		return nil, fmt.Errorf("chunkserver %s not connected", serverID)
	}

	resp, err := client.ReadChunk(ctx, &pb.ReadChunkRequest{
		Handle: handle,
	})
	if err != nil {
		return nil, fmt.Errorf("ReadChunk RPC failed: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("read failed: %s", resp.Error)
	}

	return resp.Data, nil
}

// ReplicateChunk replicates data to secondary servers via the primary
func (gc *GRPCClient) ReplicateChunk(ctx context.Context, primaryID string, secondaryIDs []string, handle string, data []byte) error {
	client, ok := gc.chunkClients[primaryID]
	if !ok {
		return fmt.Errorf("primary server %s not connected", primaryID)
	}

	// Build list of secondary server addresses
	secondaryAddrs := make([]string, len(secondaryIDs))
	for i := range secondaryIDs {
		// Map server ID to address - this should match your chunk server setup
		// In a real system, you'd look this up from the master
		secondaryAddrs[i] = "localhost:" + strconv.Itoa(5001+i)
	}

	resp, err := client.WriteChunk(ctx, &pb.WriteChunkRequest{
		Handle:   handle,
		Data:     data,
		Replicas: secondaryAddrs,
	})
	if err != nil {
		return fmt.Errorf("replication RPC failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("replication failed: %s", resp.Error)
	}

	return nil
}

// Close closes all gRPC connections
func (gc *GRPCClient) Close() {
	if gc.masterConn != nil {
		gc.masterConn.Close()
	}
	for _, conn := range gc.chunkConnections {
		if conn != nil {
			conn.Close()
		}
	}
}
