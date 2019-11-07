package server

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"

	"github.com/sleepinggenius2/go-syslog/common/message"
	"github.com/sleepinggenius2/go-syslog/server/transport"
)

type noopFormatter struct{}

func (noopFormatter) Parse() error {
	return nil
}

func (noopFormatter) Dump() message.LogParts {
	return message.LogParts{}
}

func (noopFormatter) Location(*time.Location) {}

func (n noopFormatter) GetParser(l []byte) message.LogParser {
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

func (s *handlerCounter) Handle(logParts message.LogParts, msgLen int64, err error) {
	s.current++
	if s.current == s.expected {
		close(s.done)
	}
}

func BenchmarkDatagramNoFormatting(b *testing.B) {
	handler := &handlerCounter{expected: b.N, done: make(chan struct{})}
	reader, writer := io.Pipe()
	udp := transport.NewMockPacketTransport(handler, noopFormatter{})
	udp.SetConn(&transport.PacketConnMock{PipeReader: reader})
	server := New(udp)
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
