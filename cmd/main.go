package main

import (
	"fmt"
	"github.com/bk167465/GFS/internal/master"
)

func main() {
	m := master.NewMaster()
	client := NewClient(m)

	err := client.Upload("testfile.txt")
	if err != nil {
		panic(err)
	}

	meta, ok := m.GetFileMetadata("testfile.txt")
	if !ok {
		panic("file metadata not found")
	}

	fmt.Println("File:", meta.Filename)
	for _, chunk := range meta.Chunks {
		fmt.Println("Chunk:", chunk.Handle)
		fmt.Println("Replicas:", chunk.Locations)
	}
}
