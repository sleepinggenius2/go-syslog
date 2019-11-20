package rfc3164

import (
	"bytes"
	"os"
	"time"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

type Parser struct {
	buff       []byte
	cursor     int
	l          int
	priority   message.Priority
	version    int
	header     header
	message    rfc3164message
	location   *time.Location
	skipTag    bool
	sourceType string
	cisco      *ciscoMetadata
	ciena      *cienaMetadata
}

type header struct {
	timestamp time.Time
	hostname  string
}

type rfc3164message struct {
	tag     string
	pid     string
	content string
}

func NewParser(buff []byte) *Parser {
	return &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}
}

func (p *Parser) Location(location *time.Location) {
	p.location = location
}

func (p *Parser) Parse() error {
	tcursor := p.cursor
	pri, err := p.parsePriority()
	if err != nil {
		// RFC3164 sec 4.3.3
		p.priority = message.Priority{P: 13, F: 1, S: 5}
		p.cursor = tcursor
		content, err := p.parseContent()
		p.header.timestamp = time.Now().Round(time.Second)
		if err != message.ErrEOL {
			return err
		}
		p.message = rfc3164message{content: content}
		return nil
	}

	var hdr header
	tcursor = p.cursor
	seqId := p.parseCiscoSequenceId()
	if seqId == "" {
		hdr, err = p.parseHeader()
	} else {
		p.cisco = &ciscoMetadata{seqId: seqId}
		hdr, err = p.parseCiscoHeader()
	}

	if err == message.ErrTimestampUnknownFormat {
		// RFC3164 sec 4.3.2.
		hdr.timestamp = time.Now().Round(time.Second)
		// No tag processing should be done
		p.skipTag = true
		// Reset cursor for content read
		p.cursor = tcursor
		// Reset Cisco metadata, as the message is not properly formatted
		p.cisco = nil
	} else if err != nil {
		return err
	}

	if p.cursor < p.l && p.buff[p.cursor] == ' ' {
		p.cursor++
	}

	msg, err := p.parseMessage()
	if err != message.ErrEOL {
		return err
	}

	p.priority = pri
	p.version = message.NO_VERSION
	p.header = hdr
	p.message = msg

	return nil
}

func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (p *Parser) Dump() message.LogParts {
	parts := message.LogParts{
		Priority:  p.priority.P,
		Facility:  p.priority.F,
		Severity:  p.priority.S,
		Timestamp: p.header.timestamp,
		Hostname:  p.header.hostname,
		AppName:   p.message.tag,
		ProcID:    p.message.pid,
		Message:   p.message.content,
		Received:  time.Now(),
		Valid:     true,
	}
	if p.cisco != nil {
		if p.cisco.facility == "ASA" {
			parts.SourceType = "cisco:asa"
		} else {
			parts.SourceType = "cisco:ios"
		}
		parts.StructuredData = message.StructuredData{
			"timeQuality": message.SDParams{"isSynced": boolToString(!p.cisco.notSynced)},
		}
		if p.cisco.seqId != "" {
			parts.StructuredData["meta"] = message.SDParams{"sequenceId": p.cisco.seqId}
		}
		parts.StructuredData["syslog@9"] = message.SDParams{
			"facility":    p.cisco.facility,
			"severity_id": p.cisco.severity_id,
			"mnemonic":    p.cisco.mnemonic,
		}
		if p.cisco.category != "" {
			parts.StructuredData["syslog@9"]["category"] = p.cisco.category
		}
		if p.cisco.subfacility != "" {
			parts.StructuredData["syslog@9"]["subfacility"] = p.cisco.subfacility
		}
		if p.cisco.source != "" {
			parts.StructuredData["syslog@9"]["node_id"] = p.cisco.source
		}
	} else if p.ciena != nil {
		parts.SourceType = "ciena:saos"
		parts.StructuredData = message.StructuredData{
			"origin":      message.SDParams{"ip": p.ciena.mgmtIp},
			"syslog@6141": message.SDParams{"base_mac": p.ciena.baseMac},
		}
	} else if p.sourceType != "" {
		parts.SourceType = p.sourceType
	} else {
		parts.SourceType = "syslog"
	}
	return parts
}

func (p *Parser) parsePriority() (message.Priority, error) {
	return message.ParsePriority(p.buff, &p.cursor, p.l)
}

func (p *Parser) parseHeader() (header, error) {
	hdr := header{}
	var err error

	ts, err := p.parseTimestamp()
	if err != nil {
		return hdr, err
	}

	hostname, err := p.parseHostname()
	if err != nil {
		return hdr, err
	}

	// This should be a Ciena SAOS device
	if hostname == "[local]" || hostname == "[UTC]" {
		if hostname == "[UTC]" {
			ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), ts.Nanosecond(), time.UTC)
		}
		if p.cursor < p.l && p.buff[p.cursor] == ' ' {
			p.cursor++
		}
		hostname, err = p.parseCienaHostname()
		if err != nil {
			return hdr, err
		}
	} else if len(hostname) > 0 && hostname[len(hostname)-1] == '%' {
		// This should be a Telco Systems BiNOS device
		p.sourceType = "telco:binos"
		hostname = hostname[:len(hostname)-1]
	}

	hdr.timestamp = ts
	hdr.hostname = hostname

	return hdr, nil
}

func (p *Parser) parseMessage() (rfc3164message, error) {
	msg := rfc3164message{}
	var err error

	if !p.skipTag {
		tag, pid, err := p.parseTag()
		if err != nil {
			return msg, err
		}
		msg.tag = tag
		msg.pid = pid
	}

	if p.cisco != nil {
		p.parseCiscoSystemMessage()
	}

	content, err := p.parseContent()
	if err != message.ErrEOL {
		return msg, err
	}

	msg.content = content

	return msg, err
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func (p *Parser) parseTimestamp() (time.Time, error) {
	var ts time.Time
	var err error
	var tsFmtLen int
	var sub []byte

	tsFmts := []string{
		time.Stamp,
		time.RFC3339,
	}
	// if timestamps starts with numeric try formats with different order
	// it is more likely that timestamp is in RFC3339 format then
	if c := p.buff[p.cursor]; c > '0' && c < '9' {
		tsFmts = []string{
			time.RFC3339,
			time.Stamp,
		}
	}

	found := false
	for _, tsFmt := range tsFmts {
		tsFmtLen = len(tsFmt)

		if p.cursor+tsFmtLen > p.l {
			continue
		}

		sub = p.buff[p.cursor : tsFmtLen+p.cursor]
		ts, err = time.ParseInLocation(tsFmt, string(sub), p.location)
		if err == nil {
			found = true
			break
		}
	}

	if !found {
		p.cursor = len(time.Stamp)

		// XXX : If the timestamp is invalid we try to push the cursor one byte
		// XXX : further, in case it is a space
		if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
			p.cursor++
		}

		return ts, message.ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	p.cursor += tsFmtLen

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	return ts, nil
}

func (p *Parser) parseHostname() (string, error) {
	oldcursor := p.cursor
	hostname, err := message.ParseHostname(p.buff, &p.cursor, p.l)
	if err == nil && len(hostname) > 0 && string(hostname[len(hostname)-1]) == ":" { // not an hostname! we found a GNU implementation of syslog()
		p.cursor = oldcursor - 1
		myhostname, err := os.Hostname()
		if err == nil {
			return myhostname, nil
		}
		return "", nil
	}
	return hostname, err
}

// http://tools.ietf.org/html/rfc3164#section-4.1.3
func (p *Parser) parseTag() (string, string, error) {
	var b byte
	var endOfTag bool
	var bracketOpen bool
	var bracketClosed bool
	var tag []byte
	var pid []byte
	var err error
	var found bool

	from := p.cursor
	pidFrom := 0

	for {
		if p.cursor == p.l {
			// no tag found, reset cursor for content
			p.cursor = from
			return "", "", nil
		}

		b = p.buff[p.cursor]
		bracketOpen = (b == '[')
		bracketClosed = (b == ']')
		endOfTag = (b == ':' || b == ' ')

		if bracketOpen {
			tag = p.buff[from:p.cursor]
			found = true
			pidFrom = p.cursor + 1
		}

		if bracketClosed {
			pid = p.buff[pidFrom:p.cursor]
		}

		// Support Telco Systems' use of '%' as delimeter on BiNOS
		if endOfTag || (b == '%' && p.cursor > from) {
			if !found {
				tag = p.buff[from:p.cursor]
			}

			p.cursor++
			break
		}

		p.cursor++
	}

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	if pidFrom == 0 { // No PID found
		pid = []byte{}
	}

	return string(tag), string(pid), err
}

func (p *Parser) parseContent() (string, error) {
	if p.cursor > p.l {
		return "", message.ErrEOL
	}

	content := bytes.Trim(p.buff[p.cursor:p.l], " \000")
	p.cursor += len(content)

	return string(content), message.ErrEOL
}

func fixTimestampIfNeeded(ts *time.Time) {
	now := time.Now()
	y := ts.Year()

	if ts.Year() == 0 {
		y = now.Year()
	}

	newTs := time.Date(y, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(),
		ts.Second(), ts.Nanosecond(), ts.Location())

	*ts = newTs
}
