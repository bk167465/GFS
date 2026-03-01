package master

import (
	"fmt"
	"sync"

	"github.com/bk167465/GFS/internal/common"
)

type Master struct {
	mu         sync.Mutex
	files      map[string]common.FileMetadata
	chunkCount int
	servers    []common.ServerID
}

func NewMaster() *Master {
	return &Master{
		files:   make(map[string]common.FileMetadata),
		servers: []common.ServerID{"cs1", "cs2", "cs3"},
	}
}

func (m *Master) AllocateChunk(filename string) common.ChunkMetadata {
	m.mu.Lock()
	defer m.mu.Unlock()

	handle := common.ChunkHandle(fmt.Sprintf("chunk-%d", m.chunkCount))

	var locations []common.ServerID

	// Replica placement (round robin)
	for i := 0; i < common.ReplicationFactor; i++ {
		server := m.servers[(m.chunkCount+i)%len(m.servers)]
		locations = append(locations, server)
	}

	m.chunkCount++
	primary := locations[0]

	chunk := common.ChunkMetadata{
		Handle:    handle,
		Locations: locations,
		Primary:   primary,
		Version:   1,
	}

	fileMeta := m.files[filename]
	fileMeta.Filename = filename
	fileMeta.Chunks = append(fileMeta.Chunks, chunk)
	m.files[filename] = fileMeta

	return chunk
}

func (m *Master) GetFileMetadata(filename string) (common.FileMetadata, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta, ok := m.files[filename]
	return meta, ok
}
