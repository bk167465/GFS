package main

import (
	"fmt"
	"os"

	"github.com/bk167465/GFS/internal/chunkserver"
	"github.com/bk167465/GFS/internal/common"
	"github.com/bk167465/GFS/internal/master"
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

		primaryID := string(chunkMeta.Locations[0])
		primary := c.chunkServers[primaryID]
		if primary == nil {
			return fmt.Errorf("primary server %s not found", primaryID)
		}

		// collect secondary pointers
		var secondaries []*chunkserver.ChunkServer
		for _, sid := range chunkMeta.Locations[1:] {
			sec := c.chunkServers[string(sid)]
			if sec == nil {
				return fmt.Errorf("secondary server %s not found", sid)
			}
			secondaries = append(secondaries, sec)
		}

		// primary writes and replicates
		if err := primary.Replicate(string(chunkMeta.Handle), buffer[:n], secondaries); err != nil {
			return err
		}

		if err != nil {
			break
		}
	}

	fmt.Println("Upload with replication successful!")
	return nil
}
