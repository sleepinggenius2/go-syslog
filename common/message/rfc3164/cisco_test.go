package rfc3164

import (
	"reflect"
	"testing"
	"time"

	"github.com/sleepinggenius2/go-syslog/common/message"
)

/*
== IOS & IOS XE ==
123: Jan 02 22:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime
123: Jan 02 22:04:05 UTC: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime show-timezone
123: Jan 02 22:04:05.999: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec
123: Jan 02 22:04:05.999 UTC: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec show-timezone

123: Jan 02 15:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime
123: Jan 02 15:04:05 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime show-timezone
123: Jan 02 15:04:05.999: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec localtime
123: Jan 02 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec localtime show-timezone

123: Jan 02 2006 22:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime year
123: Jan 02 2006 22:04:05 UTC: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime show-timezone year
123: Jan 02 2006 22:04:05.999: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec year
123: Jan 02 2006 22:04:05.999 UTC: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec show-timezone year

123: Jan 02 2006 15:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime year
123: Jan 02 2006 15:04:05 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime show-timezone year
123: Jan 02 2006 15:04:05.999: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec localtime year
123: Jan 02 2006 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime msec localtime show-timezone year

123: Jan  2 2006 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - Single-digit day

== IOS XR ==
123: hostnameprefix RP/0/RSP0/CPU0:Jan 02 15:04:05 : config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime
123: hostnameprefix RP/0/RSP0/CPU0:Jan 02 15:04:05 MST: config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime show-timezone
123: hostnameprefix RP/0/RSP0/CPU0:Jan 02 15:04:05.999 : config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime msec
123: hostnameprefix RP/0/RSP0/CPU0:Jan 02 15:04:05.999 MST: config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime msec show-timezone

123: hostnameprefix RP/0/RSP0/CPU0:2006 Jan 02 15:04:05 : config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime year
123: hostnameprefix RP/0/RSP0/CPU0:2006 Jan 02 15:04:05 MST: config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime show-timezone year
123: hostnameprefix RP/0/RSP0/CPU0:2006 Jan 02 15:04:05.999 : config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime msec year
123: hostnameprefix RP/0/RSP0/CPU0:2006 Jan 02 15:04:05.999 MST: config[12345]: %MGBL-SYS-5-CONFIG_I : Configured from console by admin on vty0 (192.0.2.1) - service timestamps log datetime localtime msec show-timezone year

== logging origin-id options ==
123: hostname: Jan 02 2006 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - logging origin-id hostname
123: 192.0.2.1: Jan 02 2006 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - logging origin-id ip
123: 2001:db8::1: Jan 02 2006 15:04:05.999 MST: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1) - logging origin-id ipv6
*/

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		buff    string
		want    message.LogParts
		wantErr bool
	}{
		{
			"Without hostname",
			`<190>123: Jan 02 2006 22:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1)`,
			message.LogParts{
				Priority:  190,
				Facility:  message.FacilityLocal7,
				Severity:  message.SeverityInfo,
				Timestamp: time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC),
				StructuredData: message.StructuredData{
					"timeQuality": message.SDParams{"isSynced": "1"},
					"meta":        message.SDParams{"sequenceId": "123"},
					"syslog@9":    message.SDParams{"facility": "SYS", "mnemonic": "CONFIG_I", "severity_id": "5"},
				},
				Message:    `%SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1)`,
				SourceType: "cisco:ios",
				Valid:      true,
			},
			false,
		},
		{
			"With hostname",
			`<190>123: hostname: Jan 02 2006 22:04:05: %SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1)`,
			message.LogParts{
				Priority:  190,
				Facility:  message.FacilityLocal7,
				Severity:  message.SeverityInfo,
				Timestamp: time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC),
				Hostname:  "hostname",
				StructuredData: message.StructuredData{
					"timeQuality": message.SDParams{"isSynced": "1"},
					"meta":        message.SDParams{"sequenceId": "123"},
					"syslog@9":    message.SDParams{"facility": "SYS", "mnemonic": "CONFIG_I", "severity_id": "5"},
				},
				Message:    `%SYS-5-CONFIG_I: Configured from console by admin on vty0 (192.0.2.1)`,
				SourceType: "cisco:ios",
				Valid:      true,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.buff))
			if err := p.Parse(); (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := p.Dump()
			got.Received = time.Time{}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseCiscoHeader(t *testing.T) {
	tests := []struct {
		name    string
		buff    string
		want    header
		wantErr bool
	}{
		{
			"IOS XR",
			`hostnameprefix RP/0/RSP0/CPU0:2006 Jan 02 22:04:05: `,
			header{time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC), "hostnameprefix"},
			false,
		}, {
			"IOS XR with space",
			`hostnameprefix RP/0/RSP0/CPU0 :2006 Jan 02 22:04:05: `,
			header{time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC), "hostnameprefix"},
			false,
		}, {
			"origin-id hostname",
			`hostname: Jan 02 2006 22:04:05: `,
			header{time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC), "hostname"},
			false,
		}, {
			"origin-id ip",
			`192.0.2.1: Jan 02 2006 22:04:05: `,
			header{time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC), "192.0.2.1"},
			false,
		}, {
			"origin-id ipv6",
			`2001:db8::1: Jan 02 2006 22:04:05: `,
			header{time.Date(2006, time.January, 2, 22, 4, 5, 0, time.UTC), "2001:db8::1"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.buff))
			got, err := p.parseCiscoHeader()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseCiscoHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseCiscoHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseCiscoSequenceId(t *testing.T) {
	tests := []struct {
		name string
		buff string
		want string
	}{
		{"empty", ``, ""},
		{"non-digit", `a`, ""},
		{"missing colon", `123`, ""},
		{"valid without space", `123:`, "123"},
		{"valid with space", `123: `, "123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.buff))
			if got := p.parseCiscoSequenceId(); got != tt.want {
				t.Errorf("Parser.parseCiscoSequenceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_parseCiscoHostnameAndSource(t *testing.T) {
	tests := []struct {
		name    string
		buff    string
		want    string
		want1   string
		wantErr bool
	}{
		{"IOS XR", `hostnameprefix RP/0/RSP0/CPU0:`, "hostnameprefix", "RP/0/RSP0/CPU0", false},
		{"IOS XR with space", `hostnameprefix RP/0/RSP0/CPU0 :`, "hostnameprefix", "RP/0/RSP0/CPU0", false},
		{"origin-id hostname", `hostname: `, "hostname", "", false},
		{"origin-id ip", `192.0.2.1: `, "192.0.2.1", "", false},
		{"origin-id ipv6", `2001:db8::1: `, "2001:db8::1", "", false},
		{"system clock unset", `*Jan`, "", "", false},
		{"NTP out of sync", `.Jan`, "", "", false},
		{"short month", `Jan`, "", "", false},
		{"year", `2006`, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.buff))
			got, got1, err := p.parseCiscoHostnameAndSource()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseCiscoHostnameAndSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parser.parseCiscoHostnameAndSource() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Parser.parseCiscoHostnameAndSource() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParser_parseCiscoTimestamp(t *testing.T) {
	loc, err := time.LoadLocation("MST")
	if err != nil {
		panic(err)
	}
	yearUTC := time.Now().Year()
	yearMST := time.Now().In(loc).Year()
	tests := []struct {
		name    string
		buff    string
		want    time.Time
		wantErr bool
	}{
		{"", "", time.Time{}, true},
		{"", "Jan 02 15:04:05: ", time.Date(yearUTC, time.January, 2, 15, 4, 5, 0, time.UTC), false},
		{"", "Jan 02 2006 15:04:05: ", time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC), false},
		{"", "2006 Jan 02 15:04:05 : ", time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC), false},
		{"", "Jan 02 15:04:05 MST: ", time.Date(yearMST, time.January, 2, 15, 4, 5, 0, loc), false},
		{"", "Jan 02 2006 15:04:05 MST: ", time.Date(2006, time.January, 2, 15, 4, 5, 0, loc), false},
		{"", "2006 Jan 02 15:04:05 MST: ", time.Date(2006, time.January, 2, 15, 4, 5, 0, loc), false},
		{"", "Jan 02 15:04:05.999: ", time.Date(yearUTC, time.January, 2, 15, 4, 5, 999e6, time.UTC), false},
		{"", "Jan 02 2006 15:04:05.999: ", time.Date(2006, time.January, 2, 15, 4, 5, 999e6, time.UTC), false},
		{"", "2006 Jan 02 15:04:05.999 : ", time.Date(2006, time.January, 2, 15, 4, 5, 999e6, time.UTC), false},
		{"", "Jan 02 15:04:05.999 MST: ", time.Date(yearMST, time.January, 2, 15, 4, 5, 999e6, loc), false},
		{"", "Jan 02 2006 15:04:05.999 MST: ", time.Date(2006, time.January, 2, 15, 4, 5, 999e6, loc), false},
		{"", "2006 Jan 02 15:04:05.999 MST: ", time.Date(2006, time.January, 2, 15, 4, 5, 999e6, loc), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser([]byte(tt.buff))
			got, err := p.parseCiscoTimestamp()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.parseCiscoTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.parseCiscoTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFullDate(t *testing.T) {
	tests := []struct {
		name    string
		buff    string
		want    fullDate
		wantErr bool
	}{
		// TODO: Add test cases.
		{"empty", "", fullDate{}, true},
		{"no year space-padded", "Jan  2 ", fullDate{0, 1, 2}, false},
		{"no year zero-padded", "Jan 02 ", fullDate{0, 1, 2}, false},
		{"year last space-padded", "Jan  2 2006 ", fullDate{2006, 1, 2}, false},
		{"year last zero-padded", "Jan 02 2006 ", fullDate{2006, 1, 2}, false},
		{"year first space-padded", "2006 Jan  2 ", fullDate{2006, 1, 2}, false},
		{"year first zero-padded", "2006 Jan 02 ", fullDate{2006, 1, 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cursor int
			got, err := parseFullDate([]byte(tt.buff), &cursor, len(tt.buff))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFullDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFullDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePartialTime(t *testing.T) {
	tests := []struct {
		name    string
		buff    string
		want    partialTime
		wantErr bool
	}{
		{"empty", "", partialTime{}, true},
		{"without milliseconds", "15:04:05", partialTime{15, 4, 5, 0}, false},
		{"with milliseconds", "15:04:05.999", partialTime{15, 4, 5, 999}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cursor int
			got, err := parsePartialTime([]byte(tt.buff), &cursor, len(tt.buff))
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePartialTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePartialTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFullTime(t *testing.T) {
	loc, err := time.LoadLocation("MST")
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		buff    string
		want    fullTime
		wantErr bool
	}{
		{"empty", "", fullTime{}, true},
		{"without milliseconds without timezone", "15:04:05", fullTime{partialTime{15, 4, 5, 0}, nil}, false},
		{"with milliseconds without timezone", "15:04:05.999", fullTime{partialTime{15, 4, 5, 999}, nil}, false},
		{"without milliseconds with timezone", "15:04:05 MST", fullTime{partialTime{15, 4, 5, 0}, loc}, false},
		{"with milliseconds with timezone", "15:04:05.999 MST", fullTime{partialTime{15, 4, 5, 999}, loc}, false},
		{"without milliseconds invalid timezone", "15:04:05 1", fullTime{partialTime{15, 4, 5, 0}, nil}, false},
		{"without milliseconds without timezone colon", "15:04:05:", fullTime{partialTime{15, 4, 5, 0}, nil}, false},
		{"without milliseconds without timezone space colon", "15:04:05 :", fullTime{partialTime{15, 4, 5, 0}, nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cursor int
			got, err := parseFullTime([]byte(tt.buff), &cursor, len(tt.buff))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFullTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFullTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
