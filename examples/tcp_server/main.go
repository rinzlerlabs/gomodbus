package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus/server"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	server, err := server.NewModbusTCPServer(logger, ":8502")
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
