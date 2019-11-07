package format

import (
	"bufio"
	"math"

	"github.com/pkg/errors"

	"github.com/sleepinggenius2/go-syslog/common/message"
	"github.com/sleepinggenius2/go-syslog/common/message/rfc5424"
)

const msgLenMax = bufio.MaxScanTokenSize

var (
	msgLenMaxDigits = int(math.Log10(float64(msgLenMax))) + 1

	ErrMsgLenStartNonzero = errors.New("MSG-LEN must start with NONZERO-DIGIT")
	ErrMsgLenOnlyDigit    = errors.New("MSG-LEN must contain only DIGIT")
	ErrMsgLenTooLarge     = errors.New("MSG-LEN is too large")
	ErrNotEnoughData      = errors.New("Not enough data")
)

type RFC5425 struct{}

func (f *RFC5425) GetParser(line []byte) message.LogParser {
	return rfc5424.NewParser(line)
}

func (f *RFC5425) GetSplitFunc() bufio.SplitFunc {
	return rfc5425ScannerSplit
}

func rfc5425ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if data[0] < '1' || data[0] > '9' {
		return 0, nil, ErrMsgLenStartNonzero
	}

	var i, msgLen int = 1, int(data[0] - '0')
	for ; i < len(data) && i <= msgLenMaxDigits && msgLen <= msgLenMax; i++ {
		if data[i] == ' ' {
			end := msgLen + i + 1
			if len(data) < end {
				if atEOF {
					return 0, nil, ErrNotEnoughData
				}
				return 0, nil, nil
			}
			return end, data[i+1 : end], nil
		}
		if data[i] < '0' || data[i] > '9' {
			return 0, nil, ErrMsgLenOnlyDigit
		}
		msgLen = msgLen*10 + int(data[i]-'0')
	}

	if msgLen > msgLenMax {
		return 0, nil, ErrMsgLenTooLarge
	}

	if atEOF {
		return 0, nil, ErrNotEnoughData
	}
	return 0, nil, nil
}
