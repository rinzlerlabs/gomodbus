package main

import (
	"os"
	"os/signal"

	rtu "github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	url := "rtu:///dev/ttyUSB0?baud=19200&dataBits=8&parity=N&stopBits=2&address=91"
	server, err := rtu.NewModbusServer(logger, url)
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
