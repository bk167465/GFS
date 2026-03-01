package p2p

// HandshakeFunc ...?
type HandshakeFunc func(Peer) error

func NopHandshakeFunc(Peer) error { return nil }
