package p2p

import (
	"net"
)

// Represents any arbitary data that is being
// sent over each transport between nodes in the network.
type RPC struct {
	From    net.Addr
	Payload []byte
}
