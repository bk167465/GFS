package master

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bk167465/GFS/internal/common"
	pb "github.com/bk167465/GFS/protos/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Master struct {
	mu sync.Mutex

	files       map[string]common.FileMetadata
	chunkCount  int
	servers     []common.ServerID
	serverAddrs map[common.ServerID]string

	serverHeartbeats map[common.ServerID]time.Time
	serverChunks     map[common.ServerID][]common.ChunkHandle
}

func NewMaster() *Master {
	return &Master{
		files:            make(map[string]common.FileMetadata),
		servers:          []common.ServerID{"cs1", "cs2", "cs3"},
		serverHeartbeats: make(map[common.ServerID]time.Time),
		serverChunks:     make(map[common.ServerID][]common.ChunkHandle),
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

func (m *Master) StartFailureDetector() {
	go func() {
		for {
			time.Sleep(5 * time.Second)

			m.mu.Lock()

			for server, last := range m.serverHeartbeats {
				if time.Since(last) > 20*time.Second {
					log.Println("Server DEAD:", server)
					m.handleServerFailure(server)
				}
			}

			m.mu.Unlock()
		}
	}()
}

func (m *Master) handleServerFailure(deadServer common.ServerID) {

	// Remove from heartbeat tracking
	delete(m.serverHeartbeats, deadServer)

	lostChunks := m.serverChunks[deadServer]
	delete(m.serverChunks, deadServer)

	for _, chunk := range lostChunks {

		// Find metadata
		for filename, fileMeta := range m.files {
			for i, ch := range fileMeta.Chunks {
				if ch.Handle == chunk {

					// Remove dead server from replica list
					var newLocations []common.ServerID
					for _, loc := range ch.Locations {
						if loc != deadServer {
							newLocations = append(newLocations, loc)
						}
					}

					// If replication < 3 → add new server
					if len(newLocations) < common.ReplicationFactor {
						newServer := m.findReplacementServer(newLocations)
						if newServer != "" {
							newLocations = append(newLocations, newServer)

							// Trigger replication
							go m.triggerReplication(chunk, newLocations[0], newServer)
						}
					}

					ch.Locations = newLocations
					fileMeta.Chunks[i] = ch
					m.files[filename] = fileMeta
				}
			}
		}
	}
}

func (m *Master) findReplacementServer(
	currentReplicas []common.ServerID,
) common.ServerID {

	replicaSet := make(map[common.ServerID]bool)
	for _, r := range currentReplicas {
		replicaSet[r] = true
	}

	for _, server := range m.servers {

		// Skip if already replica
		if replicaSet[server] {
			continue
		}

		// Skip if no recent heartbeat
		last, ok := m.serverHeartbeats[server]
		if !ok {
			continue
		}

		if time.Since(last) > 20*time.Second {
			continue
		}

		return server
	}

	return ""
}

func (m *Master) triggerReplication(
	chunk common.ChunkHandle,
	source common.ServerID,
	target common.ServerID,
) {

	sourceAddr := m.serverAddress(source)
	targetAddr := m.serverAddress(target)

	conn, err := grpc.Dial(targetAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("replication dial error:", err)
		return
	}
	defer conn.Close()

	client := pb.NewChunkServiceClient(conn)

	_, err = client.ReplicateChunk(context.Background(),
		&pb.ReplicateChunkRequest{
			Handle:        string(chunk),
			SourceAddress: sourceAddr,
		})

	if err != nil {
		log.Println("replication RPC failed:", err)
		return
	}

	log.Println("Replicated chunk", chunk, "from", source, "to", target)
}

func (m *Master) serverAddress(id common.ServerID) string {
	return m.serverAddrs[id]
}
