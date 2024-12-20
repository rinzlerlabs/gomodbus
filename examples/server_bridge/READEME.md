# Modbus Server Bridge

In this example, we are using 2 Modbus TCP servers, but sharing a single handler, which allows us to bridge between 2 Modbus clients that otherwise cannot talk to each other. While we use TCP for both servers here, there is no reason we can't use RTU or ASCII for one (or more) of them. This also allows us to bridge networks in the case where we want to maintain isolation.
