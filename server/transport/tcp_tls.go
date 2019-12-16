package transport

import (
	"crypto/tls"

	"github.com/pkg/errors"
)

type TCPTLSTransport struct {
	*BaseStreamTransport
	config *tls.Config
}

func (t *TCPTLSTransport) Listen() error {
	if t.tlsPeerNameFunc == nil {
		t.tlsPeerNameFunc = defaultTlsPeerName
	}

	var err error
	t.listener, err = tls.Listen("tcp", t.addr, t.config)
	if err != nil {
		return errors.Wrap(err, "Listen TCP/TLS")
	}

	t.goAcceptConnections()

	return nil
}

func NewTCPTLS(addr string, handler Handler, config *tls.Config) *TCPTLSTransport {
	return &TCPTLSTransport{
		BaseStreamTransport: newBaseStreamTransport("tcp+tls", addr, handler, RFC5425),
		config:              config,
	}
}
