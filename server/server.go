package server

import (
	"crypto/tls"
	"errors"
	"sync"

	"github.com/sleepinggenius2/go-syslog/server/transport"
)

// A function type which gets the TLS peer name from the connection. Can return
// ok=false to terminate the connection
type TlsPeerNameFunc func(tlsConn *tls.Conn) (tlsPeer string, ok bool)

type Server struct {
	doneCh     chan struct{}
	transports []transport.Transport
	wg         sync.WaitGroup
}

// NewServer returns a new Server
func New(transports ...transport.Transport) *Server {
	return &Server{transports: transports}
}

func (s *Server) AddTransport(t transport.Transport) {
	s.transports = append(s.transports, t)
}

// Starts the server and makes all the go routines go live
func (s *Server) Start() error {
	if len(s.transports) == 0 {
		return errors.New("No transports defined")
	}

	// Server is already started
	if s.doneCh != nil {
		return nil
	}
	s.doneCh = make(chan struct{})

	for _, t := range s.transports {
		t.SetSignals(&s.wg, s.doneCh)
		if err := t.Listen(); err != nil {
			return err
		}
	}

	return nil
}

// Stop the server
func (s *Server) Stop() error {
	// Server is not started yet
	if s.doneCh == nil {
		return nil
	}
	close(s.doneCh)

	for _, t := range s.transports {
		if err := t.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Waits until the server stops
func (s *Server) Wait() {
	s.wg.Wait()
}
