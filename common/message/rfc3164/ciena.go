package rfc3164

import "github.com/sleepinggenius2/go-syslog/common/message"

type cienaMetadata struct {
	baseMac string
	mgmtIp  string
}

func (p *Parser) parseCienaHostname() (string, error) {
	from := p.cursor
	to, err := message.FindNextSpace(p.buff, from, p.l)
	if err != nil {
		return "", err
	}
	mgmtIp := string(p.buff[from : to-1])
	from = to
	to, err = message.FindNextSpace(p.buff, from, p.l)
	if err != nil {
		return "", err
	}
	baseMac := string(p.buff[from : to-1])
	from = to
	to, err = message.FindNextSpace(p.buff, from, p.l)
	if err != nil {
		return "", err
	}
	hostname := string(p.buff[from : to-1])
	p.cursor = to
	p.ciena = &cienaMetadata{baseMac: baseMac, mgmtIp: mgmtIp}
	p.skipTag = true
	return hostname, nil
}
