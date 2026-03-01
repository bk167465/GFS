package main

import (
	"github.com/bk167465/GFS/internal/master"
)

func main() {
	m := master.NewMaster()
	client := NewClient(m)

	client.Upload("testfile.txt")
}
