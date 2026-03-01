package main

import (
	"fmt"
	"github.com/bk167465/GFS/internal/chunkserver"
	"github.com/bk167465/GFS/internal/common"
	"github.com/bk167465/GFS/internal/master"
	"os"
)

type Client struct {
	master       *master.Master
	chunkServers map[string]*chunkserver.ChunkServer
}

func NewClient(m *master.Master) *Client {
	return &Client{
		master: m,
		chunkServers: map[string]*chunkserver.ChunkServer{
			"cs1": chunkserver.NewChunkServer("cs1"),
			"cs2": chunkserver.NewChunkServer("cs2"),
			"cs3": chunkserver.NewChunkServer("cs3"),
		},
	}
}

func (c *Client) Upload(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		buffer := make([]byte, common.ChunkSize)
		n, err := file.Read(buffer)

		if n == 0 {
			break
		}

		// Ask master to allocate a new chunk
		chunkMeta := c.master.AllocateChunk(filename)

		// Write to ALL replicas
		for _, serverID := range chunkMeta.Locations {
			server := c.chunkServers[string(serverID)]
			if server == nil {
				return fmt.Errorf("server %s not found", serverID)
			}

			writeErr := server.WriteChunk(string(chunkMeta.Handle), buffer[:n])
			if writeErr != nil {
				return writeErr
			}
		}

		if err != nil {
			break
		}
	}

	fmt.Println("Upload with replication successful!")
	return nil
}
