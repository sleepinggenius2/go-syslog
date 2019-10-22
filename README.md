go-syslog [![GoDoc](https://godoc.org/github.com/sleepinggenius2/go-syslog?status.svg)](https://godoc.org/github.com/sleepinggenius2/go-syslog) [![GitHub release](https://img.shields.io/github/release/sleepinggenius2/go-syslog.svg)](https://github.com/sleepinggenius2/go-syslog/releases)
==============================

Syslog server library for go, build easy your custom syslog server over UDP, TCP or Unix sockets using RFC3164, RFC6587 or RFC5424

Installation
------------

The recommended way to install go-syslog

```shell
go get github.com/sleepinggenius2/go-syslog
```

Examples
--------

How import the package

```go
import "github.com/sleepinggenius2/go-syslog"
```

Example of a basic syslog [UDP server](example/basic_udp.go):

```go
channel := make(syslog.LogPartsChannel)
handler := syslog.NewChannelHandler(channel)

server := syslog.NewServer()
server.SetFormat(syslog.RFC5424)
server.SetHandler(handler)
server.ListenUDP("0.0.0.0:514")
server.Boot()

go func(channel syslog.LogPartsChannel) {
    for logParts := range channel {
        fmt.Println(logParts)
    }
}(channel)

server.Wait()
```

License
-------

MIT, see [LICENSE](LICENSE)
