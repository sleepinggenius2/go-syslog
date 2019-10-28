package format

import (
	"bufio"
	"time"

	"github.com/sleepinggenius2/go-syslog/internal/syslogparser"
)

type LogParts struct {
	syslogparser.LogParts
	Valid bool
}

type LogParser interface {
	Parse() error
	Dump() LogParts
	Location(*time.Location)
}

type Format interface {
	GetParser([]byte) LogParser
	GetSplitFunc() bufio.SplitFunc
}

type parserWrapper struct {
	syslogparser.LogParser
}

func (w *parserWrapper) Dump() LogParts {
	return LogParts{w.LogParser.Dump(), true}
}
