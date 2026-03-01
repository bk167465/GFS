package chunkserver

import (
	"os"
	"path/filepath"
	"sync"
)

type ChunkServer struct {
	ID       string
	BasePath string
	mu       sync.Mutex
	chunks   map[string]bool // local index of stored chunk handles
}

func NewChunkServer(id string) *ChunkServer {
	path := "./data/" + id
	os.MkdirAll(path, os.ModePerm)

	cs := &ChunkServer{
		ID:       id,
		BasePath: path,
		chunks:   make(map[string]bool),
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

func (cs *ChunkServer) WriteChunk(handle string, data []byte) error {
	filePath := filepath.Join(cs.BasePath, handle)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	cs.mu.Lock()
	cs.chunks[handle] = true
	cs.mu.Unlock()
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

func (cs *ChunkServer) Replicate(handle string, data []byte, peers []*ChunkServer) error {
	if err := cs.WriteChunk(handle, data); err != nil {
		return err
	}
	for _, p := range peers {
		if err := p.WriteChunk(handle, data); err != nil {
			return err
		}
	}
	return nil
}
