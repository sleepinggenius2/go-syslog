package transport

import (
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/sleepinggenius2/go-syslog/server/format"
)

var (
	RFC3164   = &format.RFC3164{}   // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC5424   = &format.RFC5424{}   // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	RFC5425   = &format.RFC5425{}   // RFC5425: http://www.ietf.org/rfc/rfc5425.txt
	RFC6587   = &format.RFC6587{}   // RFC6587: http://www.ietf.org/rfc/rfc6587.txt - octet counting variant
	Automatic = &format.Automatic{} // Automatically identify the format
)

type TimeoutCloser interface {
	Close() error
	SetReadDeadline(t time.Time) error
}

type ScanCloser struct {
	*bufio.Scanner
	closer TimeoutCloser
}

type Transport interface {
	Close() error
	GetFormat() format.Format
	GetHandler() Handler
	Listen() error
	SetFormat(f format.Format)
	SetHandler(handler Handler)
	SetSignals(wg *sync.WaitGroup, doneCh <-chan struct{})
}

type BaseTransport struct {
	addr    string
	doneCh  <-chan struct{}
	format  format.Format
	handler Handler
	wg      *sync.WaitGroup
}

func (t *BaseTransport) GetFormat() format.Format {
	return t.format
}

func (t *BaseTransport) GetHandler() Handler {
	return t.handler
}

func (t *BaseTransport) SetFormat(f format.Format) {
	t.format = f
}

func (t *BaseTransport) SetHandler(handler Handler) {
	t.handler = handler
}

func (t *BaseTransport) SetSignals(wg *sync.WaitGroup, doneCh <-chan struct{}) {
	t.wg = wg
	t.doneCh = doneCh
}

func newBaseTransport(addr string, handler Handler, f format.Format) *BaseTransport {
	return &BaseTransport{
		addr:    addr,
		format:  f,
		handler: handler,
	}
}

func parser(f format.Format, handler Handler, line []byte, client string, tlsPeer string) {
	parser := f.GetParser(line)

	err := parser.Parse()
	logParts := parser.Dump()
	logParts.Client = client
	if logParts.Hostname == "" && (f == RFC3164 || f == Automatic) {
		logParts.Hostname, _, err = net.SplitHostPort(client)
		if err != nil {
			logParts.Hostname = client
		}
	}
	logParts.TlsPeer = tlsPeer

	handler.Handle(logParts, int64(len(line)), err)
}
