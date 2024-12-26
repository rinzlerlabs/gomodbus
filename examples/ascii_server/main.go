package main

import (
	"os"
	"os/signal"

	sp "github.com/goburrow/serial"
	ascii "github.com/rinzlerlabs/gomodbus/server/serial/ascii"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	settings := &sp.Config{
		Address:  "ascii:///dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 1,
	}
	server, err := ascii.NewModbusServer(logger, settings, 91)
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
