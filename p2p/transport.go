package p2p

// Peer is an interface that represents the remote nodes.
type Peer interface {
	Close() error
}

// Transport is anything that handles communication
// between nodes in the network. This can be of the
// for (TCP, UDP, websockets,...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
