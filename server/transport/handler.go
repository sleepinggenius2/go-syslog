package transport

import "github.com/sleepinggenius2/go-syslog/common/message"

// The handler receive every syslog entry at Handle method
type Handler interface {
	Handle(logParts message.LogParts, msgLen int64, err error)
}

type LogPartsChannel chan message.LogParts

// The ChannelHandler will send all the syslog entries into the given channel
type ChannelHandler struct {
	channel LogPartsChannel
}

// NewChannelHandler returns a new ChannelHandler
func NewChannelHandler(channel LogPartsChannel) *ChannelHandler {
	return &ChannelHandler{channel: channel}
}

// The channel to be used
func (h *ChannelHandler) SetChannel(channel LogPartsChannel) {
	h.channel = channel
}

// Syslog entry receiver
func (h *ChannelHandler) Handle(logParts message.LogParts, messageLength int64, err error) {
	h.channel <- logParts
}