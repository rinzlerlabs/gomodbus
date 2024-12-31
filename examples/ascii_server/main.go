package main

import (
	"os"
	"os/signal"

	ascii "github.com/rinzlerlabs/gomodbus/server/serial/ascii"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	url := "ascii:///dev/ttyUSB0?baud=19200&dataBits=8&parity=N&stopBits=1&address=91"
	server, err := ascii.NewModbusServer(logger, url)
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}
	defer server.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
