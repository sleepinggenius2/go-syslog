package format

import (
	"bufio"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

type Format interface {
	GetParser([]byte) message.LogParser
	GetSplitFunc() bufio.SplitFunc
}
