package master

import (
	"fmt"
	"sync"

	"github.com/bk167465/GFS/internal/common"
)

type Master struct {
	mu           sync.Mutex
	files        map[string]common.FileMetadata
	chunkMapping map[common.ChunkHandle]common.ChunkMetadata
	chunkCount   int
	servers      []common.ServerID
}

func NewMaster() *Master {
	return &Master{
		files:        make(map[string]common.FileMetadata),
		chunkMapping: make(map[common.ChunkHandle]common.ChunkMetadata),
		servers:      []common.ServerID{"cs1", "cs2", "cs3"},
	}
}

func (m *Master) CreateFile(filename string, numChunks int) []common.ChunkMetadata {
	m.mu.Lock()
	defer m.mu.Unlock()

	var chunks []common.ChunkMetadata

	for i := 0; i < numChunks; i++ {
		handle := common.ChunkHandle(fmt.Sprintf("chunk-%d", m.chunkCount))
		server := m.servers[m.chunkCount%len(m.servers)]

		meta := common.ChunkMetadata{
			Handle: handle,
			Server: server,
		}

		m.chunkMapping[handle] = meta
		chunks = append(chunks, meta)
		m.chunkCount++
	}

	m.files[filename] = common.FileMetadata{
		Filename: filename,
		Chunks:   extractHandles(chunks),
	}

	return chunks
}

func extractHandles(chunks []common.ChunkMetadata) []common.ChunkHandle {
	var handles []common.ChunkHandle
	for _, c := range chunks {
		handles = append(handles, c.Handle)
	}
	return handles
}
