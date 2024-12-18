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

type modbusClient struct {
	logger          *zap.Logger
	transport       transport.Transport
	mu              sync.Mutex
	ctx             context.Context
	responseTimeout time.Duration
	aduFromRequest  func(uint16, data.FunctionCode, data.ModbusRequest) data.ModbusFrame
}

func (m *modbusClient) sendRequestAndReadResponse(address uint16, functionCode data.FunctionCode, req data.ModbusRequest) (data.ModbusFrame, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	adu := m.aduFromRequest(address, functionCode, req)
	m.logger.Debug("Sending modbus request", zap.Object("request", adu))
	err := m.transport.WriteFrame(adu)
	if err != nil {
		return nil, err
	}
	return m.readResponse()
}

func (m *modbusClient) readResponse() (data.ModbusFrame, error) {
	ctx, cancel := context.WithTimeout(m.ctx, m.responseTimeout)
	defer cancel()
	b, e := m.transport.ReadNextFrame(ctx)
	if e != nil {
		return nil, e
	}
	return b, nil
}

func (m *modbusClient) Close() error {
	return m.transport.Close()
}

func (m *modbusClient) ReadCoils(address, offset, quantity uint16) ([]bool, error) {
	req := &data.ReadCoilsRequest{
		Offset: offset,
		Count:  quantity,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.ReadCoils, req)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("Received modbus response", zap.Object("response", adu))

	resp, err := newReadCoilsResponse(adu, quantity)
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (m *modbusClient) ReadDiscreteInputs(address, offset, quantity uint16) ([]bool, error) {
	req := &data.ReadDiscreteInputsRequest{
		Offset: offset,
		Count:  quantity,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.ReadDiscreteInputs, req)
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newReadDiscreteInputsResponse(adu, quantity)
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (m *modbusClient) ReadHoldingRegisters(address, offset, quantity uint16) ([]uint16, error) {
	req := &data.ReadHoldingRegistersRequest{
		Offset: offset,
		Count:  quantity,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.ReadHoldingRegisters, req)
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newReadHoldingRegistersResponse(adu, quantity)
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (m *modbusClient) ReadInputRegisters(address, offset, quantity uint16) ([]uint16, error) {
	req := &data.ReadInputRegistersRequest{
		Offset: offset,
		Count:  quantity,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.ReadInputRegisters, req)
	if err != nil {
		return nil, err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newReadInputRegistersResponse(adu, quantity)
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

func (m *modbusClient) WriteSingleCoil(address, offset uint16, value bool) error {
	req := &data.WriteSingleCoilRequest{
		Offset: offset,
		Value:  value,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.WriteSingleCoil, req)
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newWriteSingleCoilResponse(adu)
	if err != nil {
		return err
	}
	if resp.Offset != offset {
		return common.ErrResponseOffsetMismatch
	}
	if resp.Value != value {
		return common.ErrResponseValueMismatch
	}
	return nil
}

func (m *modbusClient) WriteSingleRegister(address, offset, value uint16) error {
	req := &data.WriteSingleRegisterRequest{
		Offset: offset,
		Value:  value,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.WriteSingleRegister, req)
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newWriteSingleRegisterResponse(adu)
	if err != nil {
		return err
	}
	if resp.Offset != offset {
		return common.ErrResponseOffsetMismatch
	}
	if resp.Value != value {
		return common.ErrResponseValueMismatch
	}
	return nil
}

func (m *modbusClient) WriteMultipleCoils(address, offset uint16, values []bool) error {
	req := &data.WriteMultipleCoilsRequest{
		Offset: offset,
		Values: values,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.WriteMultipleCoils, req)
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newWriteMultipleCoilsResponse(adu)
	if err != nil {
		return err
	}
	if resp.Offset != offset {
		return common.ErrResponseOffsetMismatch
	}
	if int(resp.Count) != len(values) {
		return common.ErrResponseValueMismatch
	}
	return nil
}

func (m *modbusClient) WriteMultipleRegisters(address, offset uint16, values []uint16) error {
	req := &data.WriteMultipleRegistersRequest{
		Offset: offset,
		Values: values,
	}
	adu, err := m.sendRequestAndReadResponse(address, data.WriteMultipleRegisters, req)
	if err != nil {
		return err
	}
	m.logger.Debug("Received modbus response", zap.Object("response", adu))
	resp, err := newWriteMultipleRegistersResponse(adu)
	if err != nil {
		return err
	}
	if resp.Offset != offset {
		return common.ErrResponseOffsetMismatch
	}
	if int(resp.Count) != len(values) {
		return common.ErrResponseValueMismatch
	}
	return nil
}
