package transport

import (
	"net"

	"github.com/pkg/errors"
)

type TCPTransport struct {
	*BaseStreamTransport
}

func (t *TCPTransport) Listen() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.addr)
	if err != nil {
		return errors.Wrap(err, "Resolve TCP address")
	}

	t.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return errors.Wrap(err, "Listen TCP")
	}

	t.goAcceptConnections()

	return nil
}

func NewTCP(addr string, handler Handler) *TCPTransport {
	return &TCPTransport{BaseStreamTransport: newBaseStreamTransport(addr, handler, Automatic)}
}
