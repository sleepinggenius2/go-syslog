package format

import (
	"bufio"

	"github.com/sleepinggenius2/go-syslog/common/message"
	"github.com/sleepinggenius2/go-syslog/common/message/rfc3164"
)

type RFC3164 struct{}

func (f *RFC3164) GetParser(line []byte) message.LogParser {
	return rfc3164.NewParser(line)
}

func (f *RFC3164) GetSplitFunc() bufio.SplitFunc {
	return nil
}
