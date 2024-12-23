package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network/tcp"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	handler.(*server.DefaultHandler).Coils[0] = true
	handler.(*server.DefaultHandler).Coils[8] = true
	handler.(*server.DefaultHandler).Coils[15] = true
	server, err := tcp.NewModbusServerWithHandler(logger, ":502", handler)
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
