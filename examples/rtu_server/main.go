package main

import (
	"os"
	"os/signal"

	sp "github.com/goburrow/serial"
	rtu "github.com/rinzlerlabs/gomodbus/server/serial/rtu"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	settings := &sp.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 2,
	}
	server, err := rtu.NewModbusServer(logger, settings, 91)
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
