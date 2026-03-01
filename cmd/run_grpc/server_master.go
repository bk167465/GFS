package run_grpc

import (
	"flag"
	"log"
	"os"

	"github.com/bk167465/GFS/internal/master"
)

// runMasterServer starts the master server
func runMasterServer(port string) error {
	m := master.NewMaster()
	server := master.NewMasterServer(m)
	return server.StartServer(port)
}

// MasterCommand handles the master server startup
func MasterCommand() {
	masterCmd := flag.NewFlagSet("master", flag.ExitOnError)
	port := masterCmd.String("port", "5000", "Port for master server")

	// use os.Args directly: skip program name and subcommand
	masterCmd.Parse(os.Args[2:])

	log.Println("Starting Master Server...")
	if err := runMasterServer(*port); err != nil {
		log.Fatalf("Master server error: %v", err)
	}
}
