package server

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"

	"github.com/sleepinggenius2/go-syslog/server/format"
	"github.com/sleepinggenius2/go-syslog/server/transport"
)

type noopFormatter struct{}

func (noopFormatter) Parse() error {
	return nil
}

func (noopFormatter) Dump() format.LogParts {
	return format.LogParts{}
}

func (noopFormatter) Location(*time.Location) {}

func (n noopFormatter) GetParser(l []byte) format.LogParser {
	return n
}

func (n noopFormatter) GetSplitFunc() bufio.SplitFunc {
	return nil
}

type handlerCounter struct {
	expected int
	current  int
	done     chan struct{}
}

func (s *handlerCounter) Handle(logParts format.LogParts, msgLen int64, err error) {
	s.current++
	if s.current == s.expected {
		close(s.done)
	}
}

type fakePacketConn struct {
	*io.PipeReader
}

func (c *fakePacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, err = c.PipeReader.Read(b)
	return
}
func (c *fakePacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}
func (c *fakePacketConn) Close() error {
	return nil
}
func (c *fakePacketConn) LocalAddr() net.Addr {
	return nil
}
func (c *fakePacketConn) SetDeadline(t time.Time) error {
	return nil
}
func (c *fakePacketConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *fakePacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}
func (c *fakePacketConn) SetReadBuffer(bytes int) error {
	return nil
}

func BenchmarkDatagramNoFormatting(b *testing.B) {
	handler := &handlerCounter{expected: b.N, done: make(chan struct{})}
	reader, writer := io.Pipe()
	udp := transport.NewMockPacketTransport(handler, noopFormatter{})
	udp.SetConn(&fakePacketConn{PipeReader: reader})
	server := New(udp)
	defer func() {
		err := server.Stop()
		if err != nil {
			panic(err)
		}
	}()
	_ = udp.Listen()
	msg := []byte(exampleSyslog + "\n")
	b.SetBytes(int64(len(msg)))
	for i := 0; i < b.N; i++ {
		_, err := writer.Write(msg)
		if err != nil {
			panic(err)
		}
	}
	<-handler.done
}

func BenchmarkTCPNoFormatting(b *testing.B) {
	handler := &handlerCounter{expected: b.N, done: make(chan struct{})}
	tcp := transport.NewTCP("127.0.0.1:0", handler)
	tcp.SetFormat(noopFormatter{})
	server := New(tcp)
	defer func() {
		err := server.Stop()
		if err != nil {
			panic(err)
		}
	}()
	err := server.Start()
	if err != nil {
		panic(err)
	}
	conn, _ := net.DialTimeout("tcp", tcp.Addr().String(), time.Second)
	msg := []byte(exampleSyslog + "\n")
	b.SetBytes(int64(len(msg)))
	for i := 0; i < b.N; i++ {
		_, err = conn.Write(msg)
		if err != nil {
			panic(err)
		}
	}
	<-handler.done
}
