package transport

import (
	"fmt"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/sleepinggenius2/go-syslog/common/message"
	"github.com/sleepinggenius2/go-syslog/server/format"
)

func Test(t *testing.T) { TestingT(t) }

type TransportSuite struct{}

var _ = Suite(&TransportSuite{})
var exampleSyslog = "<31>Dec 26 05:08:46 hostname tag[296]: content"
var exampleSyslogNoTSTagHost = "<14>INFO     leaving (1) step postscripts"
var exampleSyslogNoPriority = "Dec 26 05:08:46 hostname test with no priority - see rfc 3164 section 4.3.3"
var exampleRFC5424Syslog = "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - 'su root' failed for lonvick on /dev/pts/8"

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

func (s *TransportSuite) TestConnectionClose(c *C) {
	handler := new(HandlerMock)
	tcp := NewMockStreamTransport(handler, RFC3164)
	conn := &ConnMock{ReadData: []byte(exampleSyslog)}
	tcp.SetListener(ListenerMock{Conn: conn})
	_ = tcp.Listen()
	tcp.Wait()
	c.Check(conn.isClosed, Equals, true)
}

func (s *TransportSuite) TestConnectionTCPKill(c *C) {
	handler := new(HandlerMock)
	tcp := NewMockStreamTransport(handler, RFC5424)
	conn := &ConnMock{ReadData: []byte(exampleSyslog)}
	tcp.SetListener(ListenerMock{Conn: conn})
	_ = tcp.Listen()
	err := tcp.Close()
	if err != nil {
		panic(err)
	}
	tcp.Wait()
	c.Check(conn.isClosed, Equals, true)
}

func (s *TransportSuite) TestTCPTimeout(c *C) {
	handler := new(HandlerMock)
	tcp := NewMockStreamTransport(handler, RFC3164)
	tcp.SetTimeout(10 * time.Millisecond)
	conn := &ConnMock{ReadData: []byte(exampleSyslog), ReturnTimeout: true}
	tcp.SetListener(ListenerMock{Conn: conn})
	c.Check(conn.isReadDeadline, Equals, false)
	_ = tcp.Listen()
	tcp.Wait()
	c.Check(conn.isReadDeadline, Equals, true)
	c.Check(handler.LastLogParts.Valid, Equals, false)
	c.Check(handler.LastMessageLength, Equals, int64(0))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDP3164(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, RFC3164)
	_ = udp.Listen()
	udp.SendMessage(exampleSyslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "hostname")
	c.Check(handler.LastLogParts.AppName, Equals, "tag")
	c.Check(handler.LastLogParts.Message, Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDP3164NoTag(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, RFC3164)
	_ = udp.Listen()
	udp.SendMessage(exampleSyslogNoTSTagHost)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "127.0.0.1")
	c.Check(handler.LastLogParts.AppName, Equals, "")
	c.Check(handler.LastLogParts.Message, Equals, "INFO     leaving (1) step postscripts")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslogNoTSTagHost)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDPAutomatic3164NoPriority(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, Automatic)
	_ = udp.Listen()
	udp.SendMessage(exampleSyslogNoPriority)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "127.0.0.1")
	c.Check(handler.LastLogParts.AppName, Equals, "")
	c.Check(handler.LastLogParts.Priority, Equals, 13)
	c.Check(handler.LastLogParts.Message, Equals, exampleSyslogNoPriority)
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslogNoPriority)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDP6587(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, RFC6587)
	_ = udp.Listen()
	framedSyslog := fmt.Sprintf("%d %s", len(exampleRFC5424Syslog), exampleRFC5424Syslog)
	udp.SendMessage(framedSyslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts.Facility, Equals, message.FacilityAuth)
	c.Check(handler.LastLogParts.Message, Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDPAutomatic3164(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, Automatic)
	_ = udp.Listen()
	udp.SendMessage(exampleSyslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "hostname")
	c.Check(handler.LastLogParts.AppName, Equals, "tag")
	c.Check(handler.LastLogParts.Message, Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDPAutomatic5424(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, Automatic)
	_ = udp.Listen()
	udp.SendMessage(exampleRFC5424Syslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts.Facility, Equals, message.FacilityAuth)
	c.Check(handler.LastLogParts.Message, Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDPAutomatic3164Plus6587OctetCount(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, Automatic)
	_ = udp.Listen()
	framedSyslog := fmt.Sprintf("%d %s", len(exampleSyslog), exampleSyslog)
	udp.SendMessage(framedSyslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "hostname")
	c.Check(handler.LastLogParts.AppName, Equals, "tag")
	c.Check(handler.LastLogParts.Message, Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *TransportSuite) TestUDPAutomatic5424Plus6587OctetCount(c *C) {
	handler := new(HandlerMock)
	udp := NewMockPacketTransport(handler, Automatic)
	_ = udp.Listen()
	framedSyslog := fmt.Sprintf("%d %s", len(exampleRFC5424Syslog), exampleRFC5424Syslog)
	udp.SendMessage(framedSyslog)
	udp.Wait()
	c.Check(handler.LastLogParts.Hostname, Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts.Facility, Equals, message.FacilityAuth)
	c.Check(handler.LastLogParts.Message, Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}
