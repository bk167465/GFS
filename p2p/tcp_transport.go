package p2p

import (
	"fmt"
	"net"
)

// TcpPerr represents the remote node over a TCP established connection
type TcpPeer struct {
	conn net.Conn

	// if we dial and retreive a connection => outbound == true
	// if we accept and retreive a connection => outbound == false
	outbound bool
}

func NewTcpPeer(conn net.Conn, outbound bool) *TcpPeer {
	return &TcpPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TcpPeer) Close() error {
	return p.conn.Close()
}

type TcpTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TcpTransport struct {
	TcpTransportOpts
	listener net.Listener
	rpcCh    chan RPC
}

func NewTcpTransport(opts TcpTransportOpts) *TcpTransport {
	return &TcpTransport{
		TcpTransportOpts: opts,
		rpcCh:            make(chan RPC),
	}
}

// This will return read-only channel for reading incoming messages
// received from another peer in the packet
func (t *TcpTransport) Consume() <-chan RPC {
	return t.rpcCh
}

func (t *TcpTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()

	return nil
}

func (t *TcpTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("new incoming connection %+v\n", conn)

		go t.handleConn(conn)
	}
}

func (t *TcpTransport) handleConn(conn net.Conn) {
	var err error

	peer := NewTcpPeer(conn, true)

	defer func() {
		fmt.Printf("dropping peer connection %s", err)
		conn.Close()
	}()

	if err = t.HandshakeFunc(peer); err != nil {
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	//Read Loop
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP decode error: %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcCh <- rpc
	}
}
