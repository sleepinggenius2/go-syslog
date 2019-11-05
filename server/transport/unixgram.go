package transport

import (
	"net"

	"github.com/pkg/errors"
)

type UnixgramTransport struct {
	*BasePacketTransport
}

func (t *UnixgramTransport) Listen() error {
	unixAddr, err := net.ResolveUnixAddr("unixgram", t.addr)
	if err != nil {
		return errors.Wrap(err, "Resolve Unixgram address")
	}

	t.conn, err = net.ListenUnixgram("unixgram", unixAddr)
	if err != nil {
		return errors.Wrap(err, "Listen Unixgram")
	}

	err = t.conn.SetReadBuffer(t.readBufferSize)
	if err != nil {
		return errors.Wrap(err, "Set read buffer")
	}

	t.goReceivePackets()

	return nil
}

func NewUnixgram(addr string, handler Handler) *UnixgramTransport {
	return &UnixgramTransport{
		BasePacketTransport: newBasePacketTransport(addr, handler, Automatic),
	}
}
