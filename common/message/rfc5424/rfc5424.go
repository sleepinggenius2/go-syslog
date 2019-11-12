// Note to self : never try to code while looking after your kids
// The result might look like this : https://pbs.twimg.com/media/BXqSuYXIEAAscVA.png

package rfc5424

import (
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

var (
	ErrYearInvalid       = &message.ParserError{ErrorString: "Invalid year in timestamp"}
	ErrMonthInvalid      = &message.ParserError{ErrorString: "Invalid month in timestamp"}
	ErrDayInvalid        = &message.ParserError{ErrorString: "Invalid day in timestamp"}
	ErrHourInvalid       = &message.ParserError{ErrorString: "Invalid hour in timestamp"}
	ErrMinuteInvalid     = &message.ParserError{ErrorString: "Invalid minute in timestamp"}
	ErrSecondInvalid     = &message.ParserError{ErrorString: "Invalid second in timestamp"}
	ErrSecFracInvalid    = &message.ParserError{ErrorString: "Invalid fraction of second in timestamp"}
	ErrTimeZoneInvalid   = &message.ParserError{ErrorString: "Invalid time zone in timestamp"}
	ErrInvalidTimeFormat = &message.ParserError{ErrorString: "Invalid time format"}
	ErrInvalidAppName    = &message.ParserError{ErrorString: "Invalid app name"}
	ErrInvalidProcId     = &message.ParserError{ErrorString: "Invalid proc ID"}
	ErrInvalidMsgId      = &message.ParserError{ErrorString: "Invalid msg ID"}
	ErrNoStructuredData  = &message.ParserError{ErrorString: "No structured data"}
)

type Parser struct {
	buff           []byte
	cursor         int
	l              int
	header         header
	structuredData message.StructuredData
	message        string
}

type header struct {
	priority  message.Priority
	version   int
	timestamp time.Time
	hostname  string
	appName   string
	procId    string
	msgId     string
}

type partialTime struct {
	hour    int
	minute  int
	seconds int
	secFrac float64
}

type fullTime struct {
	pt  partialTime
	loc *time.Location
}

type fullDate struct {
	year  int
	month int
	day   int
}

func NewParser(buff []byte) *Parser {
	return &Parser{
		buff:   buff,
		cursor: 0,
		l:      len(buff),
	}
}

func (p *Parser) Location(location *time.Location) {
	// Ignore as RFC5424 syslog always has a timezone
}

func (p *Parser) Parse() error {
	hdr, err := p.parseHeader()
	if err != nil {
		return err
	}

	p.header = hdr

	sd, err := p.parseStructuredData()
	if err != nil {
		return err
	}

	p.structuredData = sd
	p.cursor++

	if p.cursor < p.l {
		p.message = string(p.buff[p.cursor:])
	}

	return nil
}

func (p *Parser) Dump() message.LogParts {
	return message.LogParts{
		Priority:       p.header.priority.P,
		Facility:       p.header.priority.F,
		Severity:       p.header.priority.S,
		Version:        p.header.version,
		Timestamp:      p.header.timestamp,
		Hostname:       p.header.hostname,
		AppName:        p.header.appName,
		ProcID:         p.header.procId,
		MsgID:          p.header.msgId,
		StructuredData: p.structuredData,
		Message:        p.message,
		Received:       time.Now(),
		Valid:          true,
	}
}

// HEADER = PRI VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID
func (p *Parser) parseHeader() (header, error) {
	hdr := header{}

	pri, err := p.parsePriority()
	if err != nil {
		return hdr, err
	}

	hdr.priority = pri

	ver, err := p.parseVersion()
	if err != nil {
		return hdr, err
	}
	hdr.version = ver
	p.cursor++

	ts, err := p.parseTimestamp()
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	p.cursor++

	host, err := p.parseHostname()
	if err != nil {
		return hdr, err
	}

	hdr.hostname = host
	p.cursor++

	appName, err := p.parseAppName()
	if err != nil {
		return hdr, err
	}

	hdr.appName = appName
	p.cursor++

	procId, err := p.parseProcId()
	if err != nil {
		return hdr, nil
	}

	hdr.procId = procId
	p.cursor++

	msgId, err := p.parseMsgId()
	if err != nil {
		return hdr, nil
	}

	hdr.msgId = msgId
	p.cursor++

	return hdr, nil
}

func (p *Parser) parsePriority() (message.Priority, error) {
	return message.ParsePriority(p.buff, &p.cursor, p.l)
}

func (p *Parser) parseVersion() (int, error) {
	return message.ParseVersion(p.buff, &p.cursor, p.l)
}

// https://tools.ietf.org/html/rfc5424#section-6.2.3
func (p *Parser) parseTimestamp() (time.Time, error) {
	var ts time.Time

	if p.cursor >= p.l {
		return ts, ErrInvalidTimeFormat
	}

	if p.buff[p.cursor] == message.NILVALUE {
		p.cursor++
		return ts, nil
	}

	fd, err := parseFullDate(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, err
	}

	if p.cursor >= p.l || p.buff[p.cursor] != 'T' {
		return ts, ErrInvalidTimeFormat
	}

	p.cursor++

	ft, err := parseFullTime(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, message.ErrTimestampUnknownFormat
	}

	nSec, err := toNSec(ft.pt.secFrac)
	if err != nil {
		return ts, err
	}

	ts = time.Date(
		fd.year,
		time.Month(fd.month),
		fd.day,
		ft.pt.hour,
		ft.pt.minute,
		ft.pt.seconds,
		nSec,
		ft.loc,
	)

	return ts, nil
}

// HOSTNAME = NILVALUE / 1*255PRINTUSASCII
func (p *Parser) parseHostname() (string, error) {
	return message.ParseHostname(p.buff, &p.cursor, p.l)
}

// APP-NAME = NILVALUE / 1*48PRINTUSASCII
func (p *Parser) parseAppName() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 48, ErrInvalidAppName)
}

// PROCID = NILVALUE / 1*128PRINTUSASCII
func (p *Parser) parseProcId() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 128, ErrInvalidProcId)
}

// MSGID = NILVALUE / 1*32PRINTUSASCII
func (p *Parser) parseMsgId() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 32, ErrInvalidMsgId)
}

func (p *Parser) parseStructuredData() (message.StructuredData, error) {
	return parseStructuredData(p.buff, &p.cursor, p.l)
}

// ----------------------------------------------
// https://tools.ietf.org/html/rfc5424#section-6
// ----------------------------------------------

// XXX : bind them to Parser ?

// FULL-DATE : DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
func parseFullDate(buff []byte, cursor *int, l int) (fullDate, error) {
	var fd fullDate

	year, err := parseYear(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if *cursor >= l || buff[*cursor] != '-' {
		return fd, message.ErrTimestampUnknownFormat
	}

	*cursor++

	month, err := parseMonth(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if *cursor >= l || buff[*cursor] != '-' {
		return fd, message.ErrTimestampUnknownFormat
	}

	*cursor++

	day, err := parseDay(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	fd = fullDate{
		year:  year,
		month: month,
		day:   day,
	}

	return fd, nil
}

// DATE-FULLYEAR   = 4DIGIT
func parseYear(buff []byte, cursor *int, l int) (int, error) {
	yearLen := 4

	if *cursor+yearLen > l {
		return 0, message.ErrEOL
	}

	// XXX : we do not check for a valid year (ie. 1999, 2013 etc)
	// XXX : we only checks the format is correct
	sub := string(buff[*cursor : *cursor+yearLen])

	*cursor += yearLen

	year, err := strconv.Atoi(sub)
	if err != nil {
		return 0, ErrYearInvalid
	}

	return year, nil
}

// DATE-MONTH = 2DIGIT  ; 01-12
func parseMonth(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 1, 12, ErrMonthInvalid)
}

// DATE-MDAY = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on month/year
func parseDay(buff []byte, cursor *int, l int) (int, error) {
	// XXX : this is a relaxed constraint
	// XXX : we do not check if valid regarding February or leap years
	// XXX : we only checks that day is in range [01 -> 31]
	// XXX : in other words this function will not rant if you provide Feb 31th
	return message.Parse2Digits(buff, cursor, l, 1, 31, ErrDayInvalid)
}

// FULL-TIME = PARTIAL-TIME TIME-OFFSET
func parseFullTime(buff []byte, cursor *int, l int) (fullTime, error) {
	var ft fullTime

	pt, err := parsePartialTime(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	loc, err := parseTimeOffset(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	ft = fullTime{
		pt:  pt,
		loc: loc,
	}

	return ft, nil
}

// PARTIAL-TIME = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND[TIME-SECFRAC]
func parsePartialTime(buff []byte, cursor *int, l int) (partialTime, error) {
	var pt partialTime

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return pt, ErrInvalidTimeFormat
	}

	*cursor++

	// ----

	seconds, err := parseSecond(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	pt = partialTime{
		hour:    hour,
		minute:  minute,
		seconds: seconds,
	}

	// ----

	if *cursor >= l || buff[*cursor] != '.' {
		return pt, nil
	}

	*cursor++

	secFrac, err := parseSecFrac(buff, cursor, l)
	if err != nil {
		return pt, nil
	}
	pt.secFrac = secFrac

	return pt, nil
}

// TIME-HOUR = 2DIGIT  ; 00-23
func parseHour(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 23, ErrHourInvalid)
}

// TIME-MINUTE = 2DIGIT  ; 00-59
func parseMinute(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 59, ErrMinuteInvalid)
}

// TIME-SECOND = 2DIGIT  ; 00-59
func parseSecond(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 59, ErrSecondInvalid)
}

// TIME-SECFRAC = "." 1*6DIGIT
func parseSecFrac(buff []byte, cursor *int, l int) (float64, error) {
	maxDigitLen := 6

	max := *cursor + maxDigitLen
	from := *cursor
	to := from

	for ; to < max; to++ {
		if to >= l {
			break
		}

		c := buff[to]
		if !message.IsDigit(c) {
			break
		}
	}

	sub := string(buff[from:to])
	if len(sub) == 0 {
		return 0, ErrSecFracInvalid
	}

	secFrac, err := strconv.ParseFloat("0."+sub, 64)
	*cursor = to
	if err != nil {
		return 0, ErrSecFracInvalid
	}

	return secFrac, nil
}

// TIME-OFFSET = "Z" / TIME-NUMOFFSET
func parseTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {

	if *cursor >= l || buff[*cursor] == 'Z' {
		*cursor++
		return time.UTC, nil
	}

	return parseNumericalTimeOffset(buff, cursor, l)
}

// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
func parseNumericalTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {
	var loc = new(time.Location)

	sign := buff[*cursor]

	if (sign != '+') && (sign != '-') {
		return loc, ErrTimeZoneInvalid
	}

	*cursor++

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return loc, err
	}

	tzStr := fmt.Sprintf("%s%02d:%02d", string(sign), hour, minute)
	tmpTs, err := time.Parse("-07:00", tzStr)
	if err != nil {
		return loc, err
	}

	return tmpTs.Location(), nil
}

func getHourMinute(buff []byte, cursor *int, l int) (int, int, error) {
	hour, err := parseHour(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return 0, 0, ErrInvalidTimeFormat
	}

	*cursor++

	minute, err := parseMinute(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	return hour, minute, nil
}

func toNSec(sec float64) (int, error) {
	_, frac := math.Modf(sec)
	fracStr := strconv.FormatFloat(frac, 'f', 9, 64)
	fracInt, err := strconv.Atoi(fracStr[2:])
	if err != nil {
		return 0, err
	}

	return fracInt, nil
}

// https://tools.ietf.org/html/rfc5424#section-6.3
func parseStructuredData(buff []byte, cursor *int, l int) (message.StructuredData, error) {
	// No more data
	if *cursor >= l {
		return nil, nil
	}

	// Check for proper NILVALUE
	if buff[*cursor] == message.NILVALUE {
		*cursor++
		if *cursor < l && buff[*cursor] != ' ' {
			return nil, ErrNoStructuredData
		}
		return nil, nil
	}

	// Empty structured data (not RFC-compliant)
	if buff[*cursor] == ' ' {
		return nil, nil
	}

	// Check that there is a starting open bracket
	if buff[*cursor] != '[' {
		return nil, ErrNoStructuredData
	}

	out := make(message.StructuredData)
	from, to := *cursor, *cursor

	var (
		offset, size                              int
		inElement, inID, inParam, inName, inValue bool
		currID                                    message.SDID
		currName                                  string
		ch                                        rune
	)
loop:
	for ; to < l-offset; to += size {
		if inValue {
			ch, size = utf8.DecodeRune(buff[to+offset:])
			if offset > 0 {
				copy(buff[to:], buff[to+offset:to+offset+size])
			}
		} else {
			ch, size = rune(buff[to]), 1
		}
		switch ch {
		case '\\':
			if !inValue {
				return nil, errors.New("Invalid '\\'")
			}
			if to < l-1 {
				switch buff[to+1] {
				case '"', '\\', ']':
					offset++
					if to < l-offset {
						buff[to] = buff[to+offset]
					}
				}
			}
		case '[':
			switch {
			case inID || inName || inValue:
				break
			case inElement:
				return nil, errors.New("Invalid '['")
			default:
				inElement = true
				inID = true
				from = to + 1
			}
		case ']':
			switch {
			case inID && to > from:
				currID = message.SDID(buff[from:to])
				out[currID] = make(message.SDParams)
				inID = false
				inElement = false
			case inValue && currName != "":
				return nil, errors.New("Must escape ']' inside of PARAM-VALUE")
			case inParam:
				if from == to {
					return nil, errors.New("Missing SD-PARAM")
				}
				if inName || currName != "" {
					return nil, errors.New("Missing PARAM-VALUE")
				}
				fallthrough
			case inElement:
				if from == to {
					return nil, errors.New("Empty SD-ELEMENT")
				}
				currID = ""
				inParam = false
				inElement = false
			default:
				return nil, errors.New("Cannot have ']' outside of SD-ELEMENT")
			}
		case '=':
			switch {
			case inValue:
				break
			case inName:
				currName = string(buff[from:to])
				inName = false
				from = to + 1
			case inID:
				return nil, errors.New("SD-ID cannot contain '='")
			case !inElement:
				return nil, errors.New("Cannot have '=' outside of SD-ELEMENT")
			}
		case ' ':
			switch {
			case inID:
				if from == to {
					return nil, errors.New("Missing SD-ID")
				}
				currID = message.SDID(buff[from:to])
				out[currID] = make(message.SDParams)
				inID = false
				inParam = true
				inName = true
				from = to + 1
			case inName:
				return nil, errors.New("PARAM-NAME cannot contain ' '")
			case inValue:
				break
			case inParam:
				if currName != "" {
					return nil, errors.New("Missing PARAM-VALUE")
				}
				inName = true
				from = to + 1
			case !inElement:
				break loop
			}
		case '"':
			switch {
			case inName:
				return nil, errors.New("PARAM-NAME cannot contain '\"'")
			case inID:
				return nil, errors.New("SD-ID cannot contain '\"'")
			case inValue:
				out[currID][currName] = string(buff[from:to])
				currName = ""
				inValue = false
				to += offset
				offset = 0
			case inParam:
				inValue = true
				from = to + 1
			case !inElement:
				return nil, errors.New("Cannot have '\"' outside of SD-ELEMENT")
			}
		default:
			if !inElement {
				return nil, errors.New("Invalid character outside of SD-ELEMENT")
			}
			if inID {
				// Check for 1*32PRINTUSASCII
				if to-from == 32 {
					return nil, errors.New("SD-ID length must be <= 32")
				}
				// Check for PRINTUSASCII = %d33-126
				if buff[to] < 33 || buff[to] > 126 {
					return nil, errors.New("SD-ID must contain only PRINTUSASCII characters")
				}
			}
			if inName {
				// Check for 1*32PRINTUSASCII
				if to-from == 32 {
					return nil, errors.New("PARAM-NAME length must be <= 32")
				}
				// Check for PRINTUSASCII = %d33-126
				if buff[to] < 33 || buff[to] > 126 {
					return nil, errors.New("PARAM-NAME must contain only PRINTUSASCII characters")
				}
			}
		}
	}
	if inElement {
		return nil, errors.New("Unterminated SD-ELEMENT")
	}
	*cursor = to
	return out, nil
}

func parseUpToLen(buff []byte, cursor *int, l int, maxLen int, e error) (string, error) {
	var to int
	var found bool
	var result string

	max := *cursor + maxLen

	for to = *cursor; (to <= max) && (to < l); to++ {
		if buff[to] == ' ' {
			found = true
			break
		}
	}

	if found {
		result = string(buff[*cursor:to])
	} else if to > max {
		to = max // don't go past max
	}

	*cursor = to

	if found {
		return result, nil
	}

	return "", e
}
