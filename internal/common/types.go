package common

type ChunkHandle string
type ServerID string

const ChunkSize = 1 * 1024 * 1024 // 1MB for testing
const ReplicationFactor = 3

type FileMetadata struct {
	Filename string
	Chunks   []ChunkMetadata
}

type ChunkMetadata struct {
	Handle    ChunkHandle
	Locations []ServerID
	Version   int
}
