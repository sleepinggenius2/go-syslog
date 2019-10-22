package main

import (
	"fmt"

	"github.com/sleepinggenius2/go-syslog"
)

func main() {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	err := server.ListenUDP("0.0.0.0:514")
	if err != nil {
		panic(err)
	}
	err = server.ListenTCP("0.0.0.0:514")
	if err != nil {
		panic(err)
	}

	err = server.Boot()
	if err != nil {
		panic(err)
	}

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			fmt.Println(logParts)
		}
	}(channel)

	server.Wait()
}
