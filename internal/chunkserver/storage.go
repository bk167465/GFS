package chunkserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/bk167465/GFS/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChunkServer struct {
	ID       string
	BasePath string

	mu sync.Mutex

	chunks      map[string]bool
	writeBuffer map[string][]byte
}

func NewChunkServer(id string) *ChunkServer {
	path := "./data/" + id
	os.MkdirAll(path, os.ModePerm)

	cs := &ChunkServer{
		ID:          id,
		BasePath:    path,
		chunks:      make(map[string]bool),
		writeBuffer: make(map[string][]byte),
	}

	_ = cs.loadExisting()
	return cs
}

func (cs *ChunkServer) loadExisting() error {
	files, err := os.ReadDir(cs.BasePath)
	if err != nil {
		return err
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for _, f := range files {
		if !f.IsDir() {
			cs.chunks[f.Name()] = true
		}
	}
	return nil
}

func (cs *ChunkServer) ReadChunk(handle string) ([]byte, error) {
	filePath := filepath.Join(cs.BasePath, handle)
	return os.ReadFile(filePath)
}

func (cs *ChunkServer) HasChunk(handle string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	_, ok := cs.chunks[handle]
	return ok
}

func (cs *ChunkServer) BufferWrite(dataID string, data []byte) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.writeBuffer[dataID] = data
}

func (cs *ChunkServer) ApplyWrite(handle string, dataID string, offset int64) error {
	cs.mu.Lock()
	data, exists := cs.writeBuffer[dataID]
	if !exists {
		cs.mu.Unlock()
		return os.ErrNotExist
	}
	delete(cs.writeBuffer, dataID)
	cs.mu.Unlock()

	filePath := filepath.Join(cs.BasePath, handle)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteAt(data, offset)
	if err != nil {
		return err
	}

	cs.mu.Lock()
	cs.chunks[handle] = true
	cs.mu.Unlock()

	return nil
}

func (cs *ChunkServer) startHeartbeatLoop(masterAddr string) {
	for {
		time.Sleep(10 * time.Second)

		conn, err := grpc.Dial(masterAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println("Heartbeat dial error:", err)
			continue
		}

		client := pb.NewMasterServiceClient(conn)

		_, err = client.Heartbeat(context.Background(),
			&pb.HeartbeatRequest{
				ServerId: cs.ID,
				Chunks:   cs.getAllChunks(),
			})

		if err != nil {
			log.Println("Heartbeat RPC error:", err)
		} else {
			log.Println("Heartbeat sent from", cs.ID)
		}

		conn.Close()
	}
}

func (cs *ChunkServer) storeTempData(data []byte) string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	dataID := fmt.Sprintf("replica-%d", time.Now().UnixNano())

	cs.writeBuffer[dataID] = data

	return dataID
}

func (cs *ChunkServer) getAllChunks() []string {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var handles []string
	for h := range cs.chunks {
		handles = append(handles, h)
	}
	return handles
}
