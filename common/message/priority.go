package message

import (
	"fmt"
)

type Facility int

func (x Facility) String() string {
	switch x {
	case FacilityKern:
		return "kern"
	case FacilityUser:
		return "user"
	case FacilityMail:
		return "mail"
	case FacilityDaemon:
		return "daemon"
	case FacilityAuth:
		return "auth"
	case FacilitySyslog:
		return "syslog"
	case FacilityLpr:
		return "lpr"
	case FacilityNews:
		return "news"
	case FacilityUucp:
		return "uucp"
	case FacilityCron:
		return "cron"
	case FacilityAuthpriv:
		return "authpriv"
	case FacilityFtp:
		return "ftp"
	case FacilityNtp:
		return "ntp"
	case FacilitySecurity:
		return "security"
	case FacilityConsole:
		return "console"
	case FacilitySolarisCron:
		return "solaris-cron"
	case FacilityLocal0:
		return "local0"
	case FacilityLocal1:
		return "local1"
	case FacilityLocal2:
		return "local2"
	case FacilityLocal3:
		return "local3"
	case FacilityLocal4:
		return "local4"
	case FacilityLocal5:
		return "local5"
	case FacilityLocal6:
		return "local6"
	case FacilityLocal7:
		return "local7"
	}
	return fmt.Sprintf("unknown(%d)", x)
}

const (
	FacilityKern        Facility = 0  // Kernel messages
	FacilityUser        Facility = 1  // User-level messages
	FacilityMail        Facility = 2  // Mail system
	FacilityDaemon      Facility = 3  // System daemons
	FacilityAuth        Facility = 4  // Security/authentication messages
	FacilitySyslog      Facility = 5  // Messages generated internally by syslogd
	FacilityLpr         Facility = 6  // Line printer subsystem
	FacilityNews        Facility = 7  // Network news subsystem
	FacilityUucp        Facility = 8  // UUCP subsystem
	FacilityCron        Facility = 9  // Clock daemon
	FacilityAuthpriv    Facility = 10 // Security/authentication messages
	FacilityFtp         Facility = 11 // FTP daemon
	FacilityNtp         Facility = 12 // NTP Subsystem
	FacilitySecurity    Facility = 13 // Log audit
	FacilityConsole     Facility = 14 // Log alert
	FacilitySolarisCron Facility = 15 // Scheduling daemon
	FacilityLocal0      Facility = 16 // Locally used facility 0
	FacilityLocal1      Facility = 17 // Locally used facility 1
	FacilityLocal2      Facility = 18 // Locally used facility 2
	FacilityLocal3      Facility = 19 // Locally used facility 3
	FacilityLocal4      Facility = 20 // Locally used facility 4
	FacilityLocal5      Facility = 21 // Locally used facility 5
	FacilityLocal6      Facility = 22 // Locally used facility 6
	FacilityLocal7      Facility = 23 // Locally used facility 7
)

type Severity int

func (x Severity) String() string {
	switch x {
	case SeverityEmerg:
		return "emerg"
	case SeverityAlert:
		return "alert"
	case SeverityCrit:
		return "crit"
	case SeverityErr:
		return "err"
	case SeverityWarning:
		return "warning"
	case SeverityNotice:
		return "notice"
	case SeverityInfo:
		return "info"
	case SeverityDebug:
		return "debug"
	}
	return fmt.Sprintf("unknown(%d)", x)
}

const (
	SeverityEmerg   Severity = 0 // System is unusable
	SeverityAlert   Severity = 1 // Action must be taken immediately
	SeverityCrit    Severity = 2 // Critical conditions
	SeverityErr     Severity = 3 // Error conditions
	SeverityWarning Severity = 4 // Warning conditions
	SeverityNotice  Severity = 5 // Normal but significant conditions
	SeverityInfo    Severity = 6 // Informational messages
	SeverityDebug   Severity = 7 // Debug-level messages
)
