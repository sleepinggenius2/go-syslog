package transport

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/sleepinggenius2/go-syslog/server/format"
)

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

type ListenerMock struct {
	Conn     net.Conn
	accepted bool
}

func (l ListenerMock) Accept() (net.Conn, error) {
	if l.accepted {
		return nil, net.UnknownNetworkError("i/o timeout")
	}
	time.Sleep(time.Second)
	l.accepted = true
	return l.Conn, nil
}

func (l ListenerMock) Close() error {
	return nil
}

func (l ListenerMock) Addr() net.Addr {
	return nil
}

type MockStreamTransport struct {
	*BaseStreamTransport
}

func (t *MockStreamTransport) Listen() error {
	t.wg = new(sync.WaitGroup)
	conn, err := t.listener.Accept()
	if err != nil {
		return err
	}
	t.goScanConnection(conn)
	return nil
}

func (t *MockStreamTransport) SetListener(listener net.Listener) {
	t.listener = listener
}

func (t *MockStreamTransport) Wait() {
	t.wg.Wait()
}

func NewMockStreamTransport(handler Handler, f format.Format) *MockStreamTransport {
	return &MockStreamTransport{
		BaseStreamTransport: newBaseStreamTransport("", handler, f),
	}
}

type PacketConnMock struct {
	*io.PipeReader
	Addr net.Addr
}

func (c *PacketConnMock) ReadFrom(b []byte) (int, net.Addr, error) {
	if c.PipeReader == nil {
		return 0, c.Addr, net.UnknownNetworkError("closed")
	}
	n, err := c.PipeReader.Read(b)
	if err != nil {
		err = net.UnknownNetworkError(err.Error())
	}
	return n, c.Addr, err
}
func (c *PacketConnMock) WriteTo(b []byte, addr net.Addr) (int, error) {
	return 0, nil
}
func (c *PacketConnMock) Close() error {
	if c.PipeReader != nil {
		return c.PipeReader.Close()
	}
	return nil
}
func (c *PacketConnMock) LocalAddr() net.Addr {
	return c.Addr
}
func (c *PacketConnMock) SetDeadline(t time.Time) error {
	return nil
}
func (c *PacketConnMock) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *PacketConnMock) SetWriteDeadline(t time.Time) error {
	return nil
}
func (c *PacketConnMock) SetReadBuffer(bytes int) error {
	return nil
}

type MockPacketTransport struct {
	*BasePacketTransport
	reader *io.PipeReader
	writer *io.PipeWriter
}

func (t *MockPacketTransport) Listen() error {
	if t.conn == nil {
		t.reader, t.writer = io.Pipe()
		t.conn = &PacketConnMock{
			PipeReader: t.reader,
			Addr: &net.UDPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 45789,
			},
		}
	}
	t.wg = new(sync.WaitGroup)
	t.goReceivePackets()
	return nil
}

func (t *MockPacketTransport) SendMessage(message string) {
	_, err := t.writer.Write([]byte(message))
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Millisecond)
	t.Close()
}

func (t *MockPacketTransport) SetConn(conn PacketConn) {
	t.conn = conn
}

func (t *MockPacketTransport) Wait() {
	t.wg.Wait()
}

func NewMockPacketTransport(handler Handler, f format.Format) *MockPacketTransport {
	return &MockPacketTransport{
		BasePacketTransport: newBasePacketTransport("", handler, f),
	}
}
