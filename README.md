# A golang implementation of Modbus

A Modbus library that supports RTU, ASCII, and TCP servers and clients. Examples can be found in the [examples](examples) directory.

_Note: RTU implementation does not follow strict timing requirements due to limitations in non-RTOS environments._

## Client

Creating a Modbus client is as simple as calling `<type>.NewModbusClient`. Replace `<type>` with the transport you want to use, ex: `tcp`, `rtu`, or `ascii`. Clients expose the standard Modbus functions. Due to how the Modbus TCP/UDP protocols work, the `address` parameter on these methods has no effect.

## Server

Creating a Modbus server is as simple as calling `NewModbusServer` on the appropriate type. Servers are intended to have a long lifetime, as such they have `Start()` and `Stop()` methods.

### Customizing the register sizes

All servers use the [`DefaultHandler`](server/handler.go#L24) with 65535 registers of each type by default. If you wish to have fewer registers, simply use the `NewModbusServerWithHandler` constructor.
```
handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
server, err := tcp.NewModbusServerWithHandler(logger, ":502", handler)
```

### Handler

All implementations of the server use the [`DefaultHandler`](server/handler.go#L24), however you can create your own handler if you the default one does not suit your needs. Simply implement the [`RequestHandler`](server/handler.go#L12) interface and use the `NewModbusServerWithHandler` constructor to pass in the new handler. While I provide the ability to write your own handler, it is not for the feint of heart.

## Examples

There are a handful of examples in the [`examples`](examples/) directory that cover most functionality. Each example has a readme with more information.

### Clients

These are straightforward examples of using the clients to connect to a Modbus Server and read/write values.

### Servers

These are straightforward examples of using the servers that a Modbus Client can connect to and read/write values.

### Bridge

This example shows how you can share a single `RequestHandler` between multiple servers to function as a bridge between 2 Modbus clients.
