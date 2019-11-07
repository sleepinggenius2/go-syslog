package transport

import (
	"bufio"
	"crypto/tls"
	"net"
	"time"

	"github.com/sleepinggenius2/go-syslog/server/format"
)

type StreamTransport interface {
	Transport
	Addr() net.Addr
	SetTimeout(milliseconds int64)
}

// A function type which gets the TLS peer name from the connection. Can return
// ok=false to terminate the connection
type TlsPeerNameFunc func(tlsConn *tls.Conn) (tlsPeer string, ok bool)

type BaseStreamTransport struct {
	*BaseTransport
	listener        net.Listener
	readTimeout     time.Duration
	tlsPeerNameFunc TlsPeerNameFunc
}

func (t BaseStreamTransport) Addr() net.Addr {
	if t.listener == nil {
		return nil
	}
	return t.listener.Addr()
}

func (t *BaseStreamTransport) Close() error {
	return t.listener.Close()
}

// Sets the connection timeout for TCP connections
func (t *BaseStreamTransport) SetTimeout(timeout time.Duration) {
	t.readTimeout = timeout
}

// Set the function that extracts a TLS peer name from the TLS connection
func (t *TCPTLSTransport) SetTlsPeerNameFunc(tlsPeerNameFunc TlsPeerNameFunc) {
	t.tlsPeerNameFunc = tlsPeerNameFunc
}

// Default TLS peer name function - returns the CN of the certificate
func defaultTlsPeerName(tlsConn *tls.Conn) (tlsPeer string, ok bool) {
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) <= 0 {
		return "", false
	}
	cn := state.PeerCertificates[0].Subject.CommonName
	return cn, true
}

func (t *BaseStreamTransport) goAcceptConnections() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-t.doneCh:
				return
			default:
			}
			connection, err := t.listener.Accept()
			if err != nil {
				continue
			}
			t.goScanConnection(connection)
		}
	}()
}

func (t *BaseStreamTransport) goScanConnection(connection net.Conn) {
	scanner := bufio.NewScanner(connection)
	if sf := t.format.GetSplitFunc(); sf != nil {
		scanner.Split(sf)
	}

	remoteAddr := connection.RemoteAddr()
	var client string
	if remoteAddr != nil {
		client = remoteAddr.String()
	}

	var tlsPeer string
	if tlsConn, ok := connection.(*tls.Conn); ok {
		// Handshake now so we get the TLS peer information
		if err := tlsConn.Handshake(); err != nil {
			connection.Close()
			return
		}
		if t.tlsPeerNameFunc != nil {
			var ok bool
			tlsPeer, ok = t.tlsPeerNameFunc(tlsConn)
			if !ok {
				connection.Close()
				return
			}
		}
	}

	scanCloser := &ScanCloser{scanner, connection}

	t.wg.Add(1)
	go t.scan(scanCloser, client, tlsPeer)
}

func (t *BaseStreamTransport) scan(scanCloser *ScanCloser, client string, tlsPeer string) {
	defer func() {
		scanCloser.closer.Close()
		t.wg.Done()
	}()
	for {
		select {
		case <-t.doneCh:
			return
		default:
		}
		if t.readTimeout > 0 {
			_ = scanCloser.closer.SetReadDeadline(time.Now().Add(t.readTimeout))
		}
		if scanCloser.Scan() {
			t.parser(scanCloser.Bytes(), client, tlsPeer)
		} else {
			return
		}
	}
}

func newBaseStreamTransport(addr string, handler Handler, f format.Format) *BaseStreamTransport {
	return &BaseStreamTransport{
		BaseTransport: newBaseTransport(addr, handler, f),
	}
}
