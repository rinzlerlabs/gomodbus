package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus"
	"github.com/tarm/serial"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	port, err := serial.OpenPort(&serial.Config{
		Name:     "/dev/ttyUSB0",
		Baud:     19200,
		Size:     8,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	})
	if err != nil {
		panic(err)
	}
	server, err := gomodbus.NewModbusASCIIServer(logger, port, 91)
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
