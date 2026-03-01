package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTcpTransport(t *testing.T) {
	opts := TcpTransportOpts{
		ListenAddr:    ":0",
		HandshakeFunc: NopHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}

	tr := NewTcpTransport(opts)
	assert.Equal(t, opts.ListenAddr, tr.ListenAddr)

}
