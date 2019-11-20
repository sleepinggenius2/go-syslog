package rfc3164

import (
	"time"

	"github.com/sleepinggenius2/go-syslog/common"
	"github.com/sleepinggenius2/go-syslog/common/message"
)

const maxTzLen = 5

type ciscoMetadata struct {
	seqId       string
	source      string
	notSynced   bool
	category    string
	facility    string
	subfacility string
	severity_id string
	mnemonic    string
}

func (p *Parser) parseCiscoSequenceId() string {
	maxDigitLen := 10

	max := p.cursor + maxDigitLen
	from := p.cursor
	to := from
	var seqId string

	for ; to < max; to++ {
		if to >= p.l {
			return ""
		}
		if !message.IsDigit(p.buff[to]) {
			break
		}
	}

	// Not a sequence ID, try to parse it as something else
	if p.buff[to] != ':' {
		return ""
	}

	// Account for Cisco ASA in EMBLEM format
	if to == from {
		p.cursor++
		return "0"
	}

	seqId = string(p.buff[from:to])

	if to+1 >= p.l || p.buff[to+1] == ' ' {
		p.cursor = to + 2
		return seqId
	}

	// Not a sequence ID, try to parse it as something else
	return ""
}

func (p *Parser) parseCiscoHeader() (header, error) {
	hdr := header{}

	hostname, source, err := p.parseCiscoHostnameAndSource()
	if err != nil {
		return hdr, err
	}

	ts, err := p.parseCiscoTimestamp()
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	hdr.hostname = hostname
	if p.cisco != nil {
		p.cisco.source = source
	}

	if p.cursor < p.l-1 && p.buff[p.cursor] == ' ' && p.buff[p.cursor+1] == '%' {
		p.skipTag = true
	}

	return hdr, nil
}

var shortMonths = common.NewStringSet("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")

func (p *Parser) parseCiscoHostnameAndSource() (string, string, error) {
	oldcursor := p.cursor

	hostname, err := message.ParseHostname(p.buff, &p.cursor, p.l)
	if err != nil {
		return hostname, "", err
	}

	// IOS( XE) with origin-id
	if hostname[len(hostname)-1] == ':' {
		if p.cursor < p.l && p.buff[p.cursor] == ' ' {
			p.cursor++
		}
		return hostname[:len(hostname)-1], "", nil
	}

	// Found reliable time indicator, so reset cursor to start parsing as timestamp
	if hostname[0] == '.' || hostname[0] == '*' {
		p.cursor = oldcursor
		return "", "", nil
	}

	// Found short month name, so reset cursor to start parsing as timestamp
	if shortMonths.Contains(hostname) {
		p.cursor = oldcursor
		return "", "", nil
	}

	// Found year, so reset cursor to start parsing as timestamp
	if len(hostname) == 4 && common.IsAllDigits([]byte(hostname)) {
		p.cursor = oldcursor
		return "", "", nil
	}

	// Only option left should be IOS XR, so extract the source as well
loop:
	for from, to := p.cursor+1, p.cursor+1; to < p.l; to++ {
		if p.buff[to] == ' ' {
			// Not the token we expected, so let it try to be parsed later
			if to == p.l-1 || p.buff[to+1] != ':' {
				break loop
			}
			p.cursor = to + 2
			return hostname, string(p.buff[from:to]), nil
		}
		if p.buff[to] == ':' {
			p.cursor = to + 1
			return hostname, string(p.buff[from:to]), nil
		}
	}

	return hostname, "", nil
}

func (p *Parser) parseCiscoTimestamp() (time.Time, error) {
	var ts time.Time

	if p.cursor >= p.l {
		return ts, message.ErrInvalidTimeFormat
	}

	// '*' = system clock not set, '.' = NTP not synced
	if p.buff[p.cursor] == '*' || p.buff[p.cursor] == '.' {
		p.cisco.notSynced = true
		p.cursor++
	}

	fd, err := parseFullDate(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, err
	}

	if p.cursor >= p.l || p.buff[p.cursor] != ' ' {
		return ts, message.ErrInvalidTimeFormat
	}

	p.cursor++

	ft, err := parseFullTime(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, message.ErrTimestampUnknownFormat
	}

	if ft.loc == nil {
		ft.loc = p.location
	}

	if fd.year == 0 {
		fd.year = time.Now().In(ft.loc).Year()
	}

	ts = time.Date(
		fd.year,
		time.Month(fd.month),
		fd.day,
		ft.pt.hour,
		ft.pt.minute,
		ft.pt.seconds,
		ft.pt.milliseconds*1e6,
		ft.loc,
	)

	return ts, nil
}

type fullDate struct {
	year  int
	month int
	day   int
}

// Jan _2 | Jan 02 | Jan _2 2006 | Jan 02 2006 | 2006 Jan _2 | 2006 Jan 02
func parseFullDate(buff []byte, cursor *int, l int) (fullDate, error) {
	var fd fullDate
	var foundYear bool
	from := *cursor

	// First token
	to, err := message.FindNextSpace(buff, from, l)
	if err != nil {
		return fd, err
	}
	to = to - 1

	token := buff[from:to]
	switch len(token) {
	// Might have found short month
	case 3:
		fd.month = shortMonths.Get(string(token))
		if fd.month == 0 {
			return fd, message.ErrMonthInvalid
		}
	// Might have found year
	case 4:
		for _, c := range token {
			if !message.IsDigit(c) {
				return fd, message.ErrYearInvalid
			}
			fd.year = fd.year*10 + int(c-'0')
		}
		foundYear = true
	// Invalid token
	default:
		return fd, message.ErrInvalidTimeFormat
	}

	from = to + 1
	// Adjust for single-digit day
	if !foundYear && from < l-1 && buff[from] == ' ' {
		from++
	}

	// Second token
	to, err = message.FindNextSpace(buff, from, l)
	if err != nil {
		return fd, err
	}
	to = to - 1

	token = buff[from:to]
	switch len(token) {
	// Might have found day
	case 1, 2:
		// Expecting short month as next token
		if foundYear {
			return fd, message.ErrMonthInvalid
		}
		for _, c := range token {
			if !message.IsDigit(c) {
				return fd, message.ErrDayInvalid
			}
			fd.day = fd.day*10 + int(c-'0')
		}
	// Might have found short month
	case 3:
		// Expecting day as next token
		if !foundYear {
			return fd, message.ErrDayInvalid
		}
		fd.month = shortMonths.Get(string(token))
		if fd.month == 0 {
			return fd, message.ErrMonthInvalid
		}
	// Invalid token
	default:
		return fd, message.ErrInvalidTimeFormat
	}

	from = to + 1

	// Missing year is valid
	if !foundYear && from+1 >= l {
		*cursor = to
		return fd, nil
	}

	// We need at least three more characters
	if from+3 > l {
		return fd, message.ErrInvalidTimeFormat
	}

	// Adjust for single-digit day
	if foundYear && buff[from] == ' ' {
		from++
	}

	// Third token (optional)
	to, err = message.FindNextSpace(buff, from, l)
	if err != nil {
		return fd, nil
	}
	to = to - 1

	token = buff[from:to]
	switch len(token) {
	// Might have found day
	case 1, 2:
		// Expecting year or time as next token
		if !foundYear {
			return fd, message.ErrTimestampUnknownFormat
		}
		for _, c := range token {
			if !message.IsDigit(c) {
				return fd, message.ErrTimestampUnknownFormat
			}
			fd.day = fd.day*10 + int(c-'0')
		}
	// Might have found year
	case 4:
		// Expecting day as next token
		if foundYear {
			return fd, message.ErrDayInvalid
		}
		for _, c := range token {
			if !message.IsDigit(c) {
				return fd, message.ErrYearInvalid
			}
			fd.year = fd.year*10 + int(c-'0')
		}
	// Out of date tokens
	default:
		*cursor = from - 1
		return fd, nil
	}

	*cursor = to
	return fd, nil
}

type partialTime struct {
	hour         int
	minute       int
	seconds      int
	milliseconds int
}

// PARTIAL-TIME = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND[TIME-SECFRAC]
func parsePartialTime(buff []byte, cursor *int, l int) (partialTime, error) {
	var pt partialTime

	hour, err := parseHour(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return pt, message.ErrInvalidTimeFormat
	}

	*cursor++

	minute, err := parseMinute(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	if *cursor >= l || buff[*cursor] != ':' {
		return pt, message.ErrInvalidTimeFormat
	}

	*cursor++

	seconds, err := parseSecond(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	pt = partialTime{
		hour:    hour,
		minute:  minute,
		seconds: seconds,
	}

	if *cursor >= l || buff[*cursor] != '.' {
		return pt, nil
	}

	*cursor++

	milliseconds, err := parseMillisecond(buff, cursor, l)
	if err != nil {
		return pt, err
	}
	pt.milliseconds = milliseconds

	return pt, nil
}

// TIME-HOUR = 2DIGIT  ; 00-23
func parseHour(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 23, message.ErrHourInvalid)
}

// TIME-MINUTE = 2DIGIT  ; 00-59
func parseMinute(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 59, message.ErrMinuteInvalid)
}

// TIME-SECOND = 2DIGIT  ; 00-59
func parseSecond(buff []byte, cursor *int, l int) (int, error) {
	return message.Parse2Digits(buff, cursor, l, 0, 59, message.ErrSecondInvalid)
}

// TIME-MILLISECOND = 3DIGIT  ; 000-999
func parseMillisecond(buff []byte, cursor *int, l int) (int, error) {
	max := *cursor + 3
	if max > l {
		return 0, message.ErrSecFracInvalid
	}

	var ms int
	var c byte
	for to := *cursor; to < max; to++ {
		c = buff[to]
		if !message.IsDigit(c) {
			return 0, message.ErrSecFracInvalid
		}
		ms = ms*10 + int(c-'0')
	}

	*cursor = max
	return ms, nil
}

type fullTime struct {
	pt  partialTime
	loc *time.Location
}

// FULL-TIME = PARTIAL-TIME [MST]
func parseFullTime(buff []byte, cursor *int, l int) (fullTime, error) {
	var ft fullTime
	var err error

	ft.pt, err = parsePartialTime(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	if *cursor >= l {
		return ft, nil
	}

	if buff[*cursor] == ' ' {
		*cursor++
		if *cursor >= l {
			return ft, nil
		}
	}

	if buff[*cursor] == ':' {
		*cursor++
		return ft, nil
	}

	if buff[*cursor] < 'A' || buff[*cursor] > 'Z' {
		return ft, nil
	}

	from, to, max := *cursor, *cursor, *cursor+maxTzLen
	for ; to < l && to < max; to++ {
		if buff[to] == ':' {
			break
		}
	}

	ft.loc, err = time.LoadLocation(string(buff[from:to]))
	if err != nil {
		return ft, err
	}
	*cursor = to + 1

	return ft, nil
}

func findNextHyphen(buff []byte, from int, l int) int {
	var to int

	for to = from; to < l; to++ {
		if buff[to] == '-' {
			to++
			return to
		}
	}

	return -1
}

func (p *Parser) parseCiscoSystemMessage() {
	if p.cisco == nil || p.cursor >= p.l || p.buff[p.cursor] != '%' {
		return
	}
	// TODO: Implement this
	from := p.cursor + 1
	to := findNextHyphen(p.buff, from, p.l)
	if to < 0 {
		return
	}
	var token []byte
loop:
	for to > 0 {
		token = p.buff[from : to-1]
		from = to
		switch {
		// IOS XR includes a category
		case p.cisco.source != "" && p.cisco.category == "":
			p.cisco.category = string(token)
		case p.cisco.facility == "":
			p.cisco.facility = string(token)
		case p.cisco.severity_id == "":
			switch {
			// Found the severity
			case len(token) == 1 && common.IsDigit(token[0]):
				p.cisco.severity_id = string(token)
				break loop
			// Empty subfacility
			case p.cisco.subfacility == "":
				p.cisco.subfacility = string(token)
			// Append to subfacility
			default:
				p.cisco.subfacility += "-" + string(token)
			}
		default:
			break loop
		}
		to = findNextHyphen(p.buff, from, p.l)
	}
	if to <= 0 {
		return
	}
	for to = from; to <= p.l; to++ {
		if to == p.l || p.buff[to] == ':' || p.buff[to] == ' ' {
			p.cisco.mnemonic = string(p.buff[from:to])
			break
		}
	}
}
