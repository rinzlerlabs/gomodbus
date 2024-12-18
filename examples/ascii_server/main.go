package main

import (
	"os"
	"os/signal"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	port, err := serial.Open(&serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 1,
	})
	if err != nil {
		panic(err)
	}
	server, err := server.NewModbusASCIIServer(logger, port, 91)
	if err != nil {
		panic(err)
	}
	err = server.Start()
	if err != nil {
		panic(err)
	}
	defer server.Stop()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
