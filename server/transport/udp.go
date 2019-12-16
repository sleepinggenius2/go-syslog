package transport

import (
	"net"

	"github.com/pkg/errors"
)

type UDPTransport struct {
	*BasePacketTransport
}

func (t *UDPTransport) Listen() error {
	udpAddr, err := net.ResolveUDPAddr("udp", t.addr)
	if err != nil {
		return errors.Wrap(err, "Resolve UDP address")
	}

	t.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return errors.Wrap(err, "Listen UDP")
	}

	err = t.conn.SetReadBuffer(t.readBufferSize)
	if err != nil {
		return errors.Wrap(err, "Set read buffer")
	}

	t.goReceivePackets()

	return nil
}

func NewUDP(addr string, handler Handler) *UDPTransport {
	return &UDPTransport{
		BasePacketTransport: newBasePacketTransport("udp", addr, handler, Automatic),
	}
}
