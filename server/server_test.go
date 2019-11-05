package server

import (
	"io"
	"net"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/sleepinggenius2/go-syslog/server/format"
	"github.com/sleepinggenius2/go-syslog/server/transport"
)

func Test(t *testing.T) { TestingT(t) }

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})
var exampleSyslog = "<31>Dec 26 05:08:46 hostname tag[296]: content"

func (s *ServerSuite) TestTailFile(c *C) {
	handler := new(HandlerMock)
	udp := transport.NewUDP("0.0.0.0:5141", handler)
	udp.SetFormat(transport.RFC3164)
	server := New(udp)
	err := server.Start()
	if err != nil {
		panic(err)
	}

	go func(server *Server) {
		time.Sleep(100 * time.Millisecond)

		serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:5141")
		con, _ := net.DialUDP("udp", nil, serverAddr)
		_, err = con.Write([]byte(exampleSyslog))
		if err != nil {
			panic(err)
		}
		time.Sleep(100 * time.Millisecond)

		err = server.Stop()
		if err != nil {
			panic(err)
		}
	}(server)
	server.Wait()

	c.Check(handler.LastLogParts.Hostname, Equals, "hostname")
	c.Check(handler.LastLogParts.AppName, Equals, "tag")
	c.Check(handler.LastLogParts.Message, Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

type HandlerMock struct {
	LastLogParts      format.LogParts
	LastMessageLength int64
	LastError         error
}

func (s *HandlerMock) Handle(logParts format.LogParts, msgLen int64, err error) {
	s.LastLogParts = logParts
	s.LastMessageLength = msgLen
	s.LastError = err
}

type ConnMock struct {
	ReadData       []byte
	ReturnTimeout  bool
	isClosed       bool
	isReadDeadline bool
}

func (c *ConnMock) Read(b []byte) (n int, err error) {
	if c.ReturnTimeout {
		return 0, net.UnknownNetworkError("i/o timeout")
	}
	if c.ReadData != nil {
		l := copy(b, c.ReadData)
		c.ReadData = nil
		return l, nil
	}
	return 0, io.EOF
}

func (c *ConnMock) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *ConnMock) Close() error {
	c.isClosed = true
	return nil
}

func (c *ConnMock) LocalAddr() net.Addr {
	return nil
}

func (c *ConnMock) RemoteAddr() net.Addr {
	return nil
}

func (c *ConnMock) SetDeadline(t time.Time) error {
	return nil
}

func (c *ConnMock) SetReadDeadline(t time.Time) error {
	c.isReadDeadline = true
	return nil
}

func (c *ConnMock) SetWriteDeadline(t time.Time) error {
	return nil
}

type handlerSlow struct {
	*handlerCounter
	contents []string
}

func (s *handlerSlow) Handle(logParts format.LogParts, msgLen int64, err error) {
	if len(s.contents) == 0 {
		time.Sleep(time.Second)
	}
	s.contents = append(s.contents, logParts.Message)
	s.handlerCounter.Handle(logParts, msgLen, err)
}

func (s *ServerSuite) TestUDPRace(c *C) {
	handler := &handlerSlow{handlerCounter: &handlerCounter{expected: 3, done: make(chan struct{})}}
	udp := transport.NewUDP("127.0.0.1:0", handler)
	udp.SetFormat(transport.Automatic)
	server := New(udp)
	err := server.Start()
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("udp", udp.LocalAddr().String())
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "1"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "2"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "3"))
	c.Assert(err, IsNil)
	conn.Close()
	<-handler.done
	c.Check(handler.contents, DeepEquals, []string{"content1", "content2", "content3"})
}

func (s *ServerSuite) TestTCPRace(c *C) {
	handler := &handlerSlow{handlerCounter: &handlerCounter{expected: 3, done: make(chan struct{})}}
	tcp := transport.NewTCP("127.0.0.1:0", handler)
	tcp.SetFormat(transport.Automatic)
	tcp.SetTimeout(10 * time.Millisecond)
	server := New(tcp)
	err := server.Start()
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", tcp.Addr().String())
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "1\n"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "2\n"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "3\n"))
	c.Assert(err, IsNil)
	conn.Close()
	<-handler.done
	c.Check(handler.contents, DeepEquals, []string{"content1", "content2", "content3"})
}
