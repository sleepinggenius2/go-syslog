package format

import (
	"bufio"

	"github.com/sleepinggenius2/go-syslog/common/message"
	"github.com/sleepinggenius2/go-syslog/common/message/rfc5424"
)

type RFC6587 struct{}

func (f *RFC6587) GetParser(line []byte) message.LogParser {
	return rfc5424.NewParser(line)
}

func (f *RFC6587) GetSplitFunc() bufio.SplitFunc {
	return rfc6587ScannerSplit
}

func rfc6587ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Non-Transparent-Framing
	if data[0] == '<' {
		return len(data), data, nil
	}

	return rfc5425ScannerSplit(data, atEOF)
}
