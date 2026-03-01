package chunkserver

import (
	"os"
	"path/filepath"
)

type ChunkServer struct {
	ID       string
	BasePath string
}

func NewChunkServer(id string) *ChunkServer {
	path := "./data/" + id
	os.MkdirAll(path, os.ModePerm)

	return &ChunkServer{
		ID:       id,
		BasePath: path,
	}
}

func (cs *ChunkServer) WriteChunk(handle string, data []byte) error {
	filePath := filepath.Join(cs.BasePath, handle)
	return os.WriteFile(filePath, data, 0644)
}

func (cs *ChunkServer) ReadChunk(handle string) ([]byte, error) {
	filePath := filepath.Join(cs.BasePath, handle)
	return os.ReadFile(filePath)
}
