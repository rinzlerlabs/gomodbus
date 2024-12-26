package client

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type ModbusClient interface {
	io.Closer
	// ReadCoils reads the status of coils in a remote device.
	ReadCoils(address, offset, quantity uint16) ([]bool, error)
	// ReadDiscreteInputs reads the status of discrete inputs in a remote device.
	ReadDiscreteInputs(address, offset, quantity uint16) ([]bool, error)
	// ReadHoldingRegisters reads the contents of holding registers in a remote device.
	ReadHoldingRegisters(address, offset, quantity uint16) ([]uint16, error)
	// ReadInputRegisters reads the contents of input registers in a remote device.
	ReadInputRegisters(address, offset, quantity uint16) ([]uint16, error)
	// WriteSingleCoil writes a single coil in a remote device.
	WriteSingleCoil(address, offset uint16, value bool) error
	// WriteSingleRegister writes a single holding register in a remote device.
	WriteSingleRegister(address, offset, value uint16) error
	// WriteMultipleCoils writes multiple coils in a remote device.
	WriteMultipleCoils(address, offset uint16, values []bool) error
	// WriteMultipleRegisters writes multiple holding registers in a remote device.
	WriteMultipleRegisters(address, offset uint16, values []uint16) error
}

func NewModbusClient(ctx context.Context, logger *zap.Logger, transport transport.Transport, requestCreator requestCreator, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		logger:          logger,
		transport:       transport,
		ctx:             ctx,
		responseTimeout: responseTimeout,
		requestCreator:  requestCreator,
	}
}

type modbusClient struct {
	logger          *zap.Logger
	transport       transport.Transport
	mu              sync.Mutex
	ctx             context.Context
	responseTimeout time.Duration
	requestCreator  requestCreator
}

type requestCreator interface {
	CreateTransaction(*transport.ModbusFrame, transport.Transport) transport.ModbusTransaction
	NewModbusFrame(uint16, *transport.ProtocolDataUnit) *transport.ModbusFrame
}

func (m *modbusClient) sendRequestAndReadResponse(address uint16, req *transport.ProtocolDataUnit) (*transport.ModbusFrame, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	frame := m.requestCreator.NewModbusFrame(address, req)
	txn := m.requestCreator.CreateTransaction(frame, m.transport)
	m.logger.Debug("Sending modbus request", zap.Object("Frame", txn.Frame()))
	return txn.Exchange(m.ctx)
}

func (m *modbusClient) Close() error {
	return m.transport.Close()
}

func (m *modbusClient) ReadCoils(address, offset, quantity uint16) ([]bool, error) {
	req := data.NewReadCoilsRequest(offset, quantity)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return nil, err
	}

	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	return adu.PDU().Operation().(*data.ReadCoilsResponse).Values(), nil
}

func (m *modbusClient) ReadDiscreteInputs(address, offset, quantity uint16) ([]bool, error) {
	req := data.NewReadDiscreteInputsRequest(offset, quantity)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))

	return adu.PDU().Operation().(*data.ReadDiscreteInputsResponse).Values(), nil
}

func (m *modbusClient) ReadHoldingRegisters(address, offset, quantity uint16) ([]uint16, error) {
	req := data.NewReadHoldingRegistersRequest(offset, quantity)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	return adu.PDU().Operation().(*data.ReadHoldingRegistersResponse).Values(), nil
}

func (m *modbusClient) ReadInputRegisters(address, offset, quantity uint16) ([]uint16, error) {
	req := data.NewReadInputRegistersRequest(offset, quantity)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	return adu.PDU().Operation().(*data.ReadInputRegistersResponse).Values(), nil
}

func (m *modbusClient) WriteSingleCoil(address, offset uint16, value bool) error {
	req := data.NewWriteSingleCoilRequest(offset, value)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp := adu.PDU().Operation().(*data.WriteSingleCoilResponse)
	if resp.Offset() != offset {
		return common.ErrResponseOffsetMismatch
	}
	if resp.Value() != value {
		return common.ErrResponseValueMismatch
	}
	return nil
}

func (m *modbusClient) WriteSingleRegister(address, offset, value uint16) error {
	req := data.NewWriteSingleRegisterRequest(offset, value)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	if resp, success := adu.PDU().Operation().(*data.WriteSingleRegisterResponse); success == false {
		return common.ErrInvalidPacket
	} else {
		if resp.Offset() != offset {
			return common.ErrResponseOffsetMismatch
		}
		if resp.Value() != value {
			return common.ErrResponseValueMismatch
		}
		return nil
	}
}

func (m *modbusClient) WriteMultipleCoils(address, offset uint16, values []bool) error {
	req := data.NewWriteMultipleCoilsRequest(offset, values)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	if resp, success := adu.PDU().Operation().(*data.WriteMultipleCoilsResponse); success == false {
		return common.ErrInvalidPacket
	} else {
		if resp.Offset() != offset {
			return common.ErrResponseOffsetMismatch
		}
		if int(resp.Count()) != len(values) {
			return common.ErrResponseValueMismatch
		}
		return nil
	}
}

func (m *modbusClient) WriteMultipleRegisters(address, offset uint16, values []uint16) error {
	req := data.NewWriteMultipleRegistersRequest(offset, values)
	adu, err := m.sendRequestAndReadResponse(address, transport.NewProtocolDataUnit(req))
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	if resp, success := adu.PDU().Operation().(*data.WriteMultipleRegistersResponse); success == false {
		return common.ErrInvalidPacket
	} else {
		if resp.Offset() != offset {
			return common.ErrResponseOffsetMismatch
		}
		if int(resp.Count()) != len(values) {
			return common.ErrResponseValueMismatch
		}
		return nil
	}
}
