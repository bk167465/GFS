package chunkserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunkServerWriteRead(t *testing.T) {
	// Create temp server
	server := NewChunkServer("test-cs")
	defer os.RemoveAll(server.BasePath)

	testData := []byte("Hello, GFS!")
	testHandle := "chunk-test-001"

	// Test Write
	err := server.WriteChunk(testHandle, testData)
	if err != nil {
		t.Fatalf("WriteChunk failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(server.BasePath, testHandle)
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("Chunk file not created: %v", err)
	}

	// Test Read
	readData, err := server.ReadChunk(testHandle)
	if err != nil {
		t.Fatalf("ReadChunk failed: %v", err)
	}

	// Verify content
	if string(readData) != string(testData) {
		t.Fatalf("Data mismatch. Expected %s, got %s", testData, readData)
	}
}
