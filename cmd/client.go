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

	stat, _ := file.Stat()
	numChunks := int(stat.Size()/common.ChunkSize) + 1

	chunks := c.master.CreateFile(filename, numChunks)

	for _, chunkMeta := range chunks {
		buffer := make([]byte, common.ChunkSize)
		n, _ := file.Read(buffer)

		server := c.chunkServers[string(chunkMeta.Server)]
		err := server.WriteChunk(string(chunkMeta.Handle), buffer[:n])
		if err != nil {
			return err
		}
	}

	fmt.Println("Upload successful!")
	return nil
}
