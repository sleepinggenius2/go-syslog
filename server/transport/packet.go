package transport

import (
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sleepinggenius2/go-syslog/server/format"
)

const (
	packetChannelSizeDefault    = 10
	packetReadBufferSizeDefault = bufio.MaxScanTokenSize
)

var packetBufferPool = sync.Pool{
	New: func() interface{} {
		buff := make([]byte, packetReadBufferSizeDefault)
		return &buff
	},
}

type PacketMessage struct {
	message *[]byte
	client  string
}

type PacketConn interface {
	net.PacketConn
	SetReadBuffer(bytes int) error
}

type PacketTransport interface {
	Transport
	LocalAddr() net.Addr
	SetPacketChannelSize(size int)
	SetReadBufferSize(bytes int) error
}

type BasePacketTransport struct {
	*BaseTransport
	conn              PacketConn
	packetChannel     chan PacketMessage
	packetChannelSize int
	readBufferSize    int
}

func (t *BasePacketTransport) Close() error {
	return t.conn.Close()
}

func (t BasePacketTransport) LocalAddr() net.Addr {
	if t.conn == nil {
		return nil
	}
	return t.conn.LocalAddr()
}

// Sets the packet channel size
func (t *BasePacketTransport) SetPacketChannelSize(size int) {
	t.packetChannelSize = size
}

// Sets the read buffer size
func (t *BasePacketTransport) SetReadBufferSize(bytes int) error {
	if bytes > bufio.MaxScanTokenSize {
		return errors.Errorf("Read buffer size cannot exceed %d bytes", bufio.MaxScanTokenSize)
	}
	t.readBufferSize = bytes
	if t.conn != nil {
		return t.conn.SetReadBuffer(bytes)
	}
	return nil
}

func (t *BasePacketTransport) goReceivePackets() {
	t.goParsePackets()

	t.wg.Add(1)
	go func() {
		defer func() {
			close(t.packetChannel)
			t.wg.Done()
		}()
		for {
			buff := packetBufferPool.Get().(*[]byte)
			n, addr, err := t.conn.ReadFrom(*buff)
			if err == nil {
				if n == 0 {
					*buff = (*buff)[:cap(*buff)]
					packetBufferPool.Put(buff)
				}
				if n > 0 {
					*buff = (*buff)[:n]
					var address string
					if addr != nil {
						address = addr.String()
					}
					select {
					case <-t.doneCh:
						return
					case t.packetChannel <- PacketMessage{buff, address}:
					}
				}
			} else {
				// there has been an error. Either the server has been killed
				// or may be getting a transitory error due to (e.g.) the
				// interface being shutdown in which case sleep() to avoid busy wait.
				netError, ok := err.(net.Error)
				if (ok) && !netError.Temporary() && !netError.Timeout() {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

func (t *BasePacketTransport) goParsePackets() {
	if t.packetChannel == nil {
		t.packetChannel = make(chan PacketMessage, t.packetChannelSize)
	}

	t.wg.Add(1)
	go func() {
		for {
			msg, ok := (<-t.packetChannel)
			if !ok {
				break
			}
			if sf := t.format.GetSplitFunc(); sf != nil {
				if _, token, err := sf(*msg.message, true); err == nil {
					t.parser(token, msg.client, "")
				}
			} else {
				t.parser(*msg.message, msg.client, "")
			}
			*msg.message = (*msg.message)[:cap(*msg.message)]
			packetBufferPool.Put(msg.message)
		}
		t.wg.Done()
	}()
}

func newBasePacketTransport(network string, addr string, handler Handler, f format.Format) *BasePacketTransport {
	return &BasePacketTransport{
		BaseTransport:     newBaseTransport(network, addr, handler, f),
		packetChannelSize: packetChannelSizeDefault,
		readBufferSize:    packetReadBufferSizeDefault,
	}
}
