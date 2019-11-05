package main

import (
	"fmt"

	"github.com/sleepinggenius2/go-syslog/server"
	"github.com/sleepinggenius2/go-syslog/server/transport"
)

func main() {
	channel := make(transport.LogPartsChannel)
	handler := transport.NewChannelHandler(channel)

	udp := transport.NewUDP("0.0.0.0:514", handler)
	udp.SetFormat(transport.RFC5424)
	tcp := transport.NewTCP("0.0.0.0:514", handler)
	syslog := server.New(udp, tcp)

	err := syslog.Start()
	if err != nil {
		panic(err)
	}

	go func(channel transport.LogPartsChannel) {
		for logParts := range channel {
			fmt.Println(logParts)
		}
	}(channel)

	syslog.Wait()
}
