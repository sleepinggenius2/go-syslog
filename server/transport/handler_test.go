package transport

import (
	. "gopkg.in/check.v1"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

type HandlerSuite struct{}

var _ = Suite(&HandlerSuite{})

func (s *HandlerSuite) TestHandle(c *C) {
	logPart := message.LogParts{Message: "foo"}

	channel := make(LogPartsChannel, 1)
	handler := NewChannelHandler(channel)
	handler.Handle(logPart, 10, nil)

	fromChan := <-channel
	c.Check(fromChan.Message, Equals, logPart.Message)
}