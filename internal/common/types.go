package common

type ChunkHandle string
type ServerID string

const ChunkSize = 1 * 1024 * 1024 // 1MB for testing

type FileMetadata struct {
	Filename string
	Chunks   []ChunkHandle
}

type ChunkMetadata struct {
	Handle  ChunkHandle
	Server  ServerID
	Version int
}
