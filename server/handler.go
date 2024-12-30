package server

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

const (
	// DefaultCoilCount is the default number of coils.
	DefaultCoilCount = 65535
	// DefaultDiscreteInputCount is the default number of discrete inputs.
	DefaultDiscreteInputCount = 65535
	// DefaultHoldingRegisterCount is the default number of holding registers.
	DefaultHoldingRegisterCount = 65535
	// DefaultInputRegisterCount is the default number of input registers.
	DefaultInputRegisterCount = 65535

	coilsFile            = "coils.dat"
	discreteInputsFile   = "discrete_inputs.dat"
	holdingRegistersFile = "holding_registers.dat"
	inputRegistersFile   = "input_registers.dat"
)

// RequestHandler is the interface that wraps the basic Modbus functions.
// TODO: Merge single and multiple requests into one.
type RequestHandler interface {
	// Handle handles a modbus transaction, this is used by the server to process incoming requests.
	Handle(op transport.ApplicationDataUnit) (*transport.ProtocolDataUnit, error)
	// ReadCoils reads the status of coils in this device.
	ReadCoils(request data.ModbusReadRequest) (response data.ModbusReadResponse[[]bool], err error)
	// ReadDiscreteInputs reads the status of discrete inputs in this device.
	ReadDiscreteInputs(request data.ModbusReadRequest) (response data.ModbusReadResponse[[]bool], err error)
	// ReadHoldingRegisters reads the contents of holding registers in this device.
	ReadHoldingRegisters(request data.ModbusReadRequest) (response data.ModbusReadResponse[[]uint16], err error)
	// ReadInputRegisters reads the contents of input registers in this device.
	ReadInputRegisters(request data.ModbusReadRequest) (response data.ModbusReadResponse[[]uint16], err error)
	// WriteSingleCoil writes a single coil in this device.
	WriteSingleCoil(equest data.ModbusWriteSingleRequest[bool]) (response *data.WriteSingleCoilResponse, err error)
	// WriteSingleRegister writes a single holding register in this device.
	WriteSingleRegister(request data.ModbusWriteSingleRequest[uint16]) (response *data.WriteSingleRegisterResponse, err error)
	// WriteMultipleCoils writes multiple coils in this device.
	WriteMultipleCoils(request data.ModbusWriteArrayRequest[[]bool]) (response *data.WriteMultipleCoilsResponse, err error)
	// WriteMultipleRegisters writes multiple holding registers in this device.
	WriteMultipleRegisters(request data.ModbusWriteArrayRequest[[]uint16]) (response *data.WriteMultipleRegistersResponse, err error)
}

// PersistableRequestHandler is the interface that wraps the basic Modbus functions and provides methods to load and save server data.
// This interface extends the RequestHandler interface, but as a result, adds slight performance penalty due to the need for locking.
type PersistableRequestHandler interface {
	RequestHandler
	// Load loads the server data from the specified path.
	Load(dataPath string) error
	// Save saves the server data to the specified path.
	Save(dataPath string) error
}

// DefaultHandler is the default implementation of the PersistableRequestHandler interface.
type DefaultHandler struct {
	logger           *zap.Logger
	mu               sync.RWMutex
	Coils            []bool
	DiscreteInputs   []bool
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

// NewDefaultHandler creates a new DefaultHandler with the specified register counts. This is a PersistableRequestHandler, which means there is some internal locking
// to provide thread safety when reading, loading, and saving data. If any of the register counts are 0, the default values are used.
func NewDefaultHandler(logger *zap.Logger, coilCount, discreteInputCount, holdingRegisterCount, inputRegisterCount uint16) PersistableRequestHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	if coilCount == 0 || discreteInputCount == 0 || holdingRegisterCount == 0 || inputRegisterCount == 0 {
		logger.Warn("Invalid count, using default values")
		coilCount = DefaultCoilCount
		discreteInputCount = DefaultDiscreteInputCount
		holdingRegisterCount = DefaultHoldingRegisterCount
		inputRegisterCount = DefaultInputRegisterCount
	}
	return &DefaultHandler{
		logger:           logger,
		Coils:            make([]bool, coilCount),
		DiscreteInputs:   make([]bool, discreteInputCount),
		HoldingRegisters: make([]uint16, holdingRegisterCount),
		InputRegisters:   make([]uint16, inputRegisterCount),
	}
}

func (h *DefaultHandler) Handle(adu transport.ApplicationDataUnit) (*transport.ProtocolDataUnit, error) {
	h.logger.Info("Request", zap.Object("ADU", adu))
	var result data.ModbusOperation
	var err error
	switch adu.PDU().FunctionCode() {
	case data.ReadCoils:
		// Read Coils
		result, err = h.ReadCoils(adu.PDU().Operation().(data.ModbusReadRequest))
	case data.ReadDiscreteInputs:
		// Read Discrete Inputs
		result, err = h.ReadDiscreteInputs(adu.PDU().Operation().(data.ModbusReadRequest))
	case data.ReadHoldingRegisters:
		// Read Holding Registers
		result, err = h.ReadHoldingRegisters(adu.PDU().Operation().(data.ModbusReadRequest))
	case data.ReadInputRegisters:
		// Read Input Registers
		result, err = h.ReadInputRegisters(adu.PDU().Operation().(data.ModbusReadRequest))
	case data.WriteSingleCoil:
		// Write Single Coil
		result, err = h.WriteSingleCoil(adu.PDU().Operation().(data.ModbusWriteSingleRequest[bool]))
	case data.WriteSingleRegister:
		// Write Single Register
		result, err = h.WriteSingleRegister(adu.PDU().Operation().(data.ModbusWriteSingleRequest[uint16]))
	case data.WriteMultipleCoils:
		// Write Multiple Coils
		result, err = h.WriteMultipleCoils(adu.PDU().Operation().(data.ModbusWriteArrayRequest[[]bool]))
	case data.WriteMultipleRegisters:
		// Write Multiple Registers
		result, err = h.WriteMultipleRegisters(adu.PDU().Operation().(data.ModbusWriteArrayRequest[[]uint16]))
	default:
		h.logger.Debug("Received packet with unknown function code", zap.Any("packet", adu))
		result = data.NewModbusOperationException(adu.PDU().FunctionCode(), data.IllegalFunction)
	}
	switch err {
	case nil:
		break
	case common.ErrIllegalDataAddress:
		h.logger.Error("Failed to handle request", zap.Error(err))
		result = data.NewModbusOperationException(adu.PDU().FunctionCode(), data.IllegalDataAddress)
	default:
		h.logger.Error("Failed to handle request", zap.Error(err))
		result = data.NewModbusOperationException(adu.PDU().FunctionCode(), data.ServerDeviceFailure)
	}
	pdu := transport.NewProtocolDataUnit(result)
	h.logger.Debug("Response", zap.Object("PDU", pdu))
	return pdu, nil
}

func getRange(offset uint16, count int) (uint16, uint16) {
	start := offset
	end := offset + uint16(count)
	return start, end
}

func (h *DefaultHandler) ReadCoils(operation data.ModbusReadRequest) (response data.ModbusReadResponse[[]bool], err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("ReadCoils", zap.Uint16("Offset", operation.Offset()), zap.Int("Count", operation.Count()))
	start, end := getRange(operation.Offset(), operation.Count())
	if int(end) > len(h.Coils) {
		return nil, common.ErrIllegalDataAddress
	}
	results := h.Coils[start:end]
	return data.NewReadCoilsResponse(results), nil
}

func (h *DefaultHandler) ReadDiscreteInputs(operation data.ModbusReadRequest) (response data.ModbusReadResponse[[]bool], err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("ReadDiscreteInputs", zap.Uint16("Offset", operation.Offset()), zap.Int("Count", operation.Count()))
	start, end := getRange(operation.Offset(), operation.Count())
	if int(end) > len(h.HoldingRegisters) {
		return nil, common.ErrIllegalDataAddress
	}
	results := h.DiscreteInputs[start:end]
	return data.NewReadDiscreteInputsResponse(results), nil
}

func (h *DefaultHandler) ReadHoldingRegisters(operation data.ModbusReadRequest) (response data.ModbusReadResponse[[]uint16], err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("ReadHoldingRegisters", zap.Uint16("Offset", operation.Offset()), zap.Int("Count", operation.Count()))
	start, end := getRange(operation.Offset(), operation.Count())
	if int(end) > len(h.HoldingRegisters) {
		return nil, common.ErrIllegalDataAddress
	}
	results := h.HoldingRegisters[start:end]
	return data.NewReadHoldingRegistersResponse(results), nil
}

func (h *DefaultHandler) ReadInputRegisters(operation data.ModbusReadRequest) (response data.ModbusReadResponse[[]uint16], err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("ReadInputRegisters", zap.Uint16("Offset", operation.Offset()), zap.Int("Count", operation.Count()))
	start, end := getRange(operation.Offset(), operation.Count())
	if int(end) > len(h.InputRegisters) {
		return nil, common.ErrIllegalDataAddress
	}
	results := h.InputRegisters[start:end]
	return data.NewReadInputRegistersResponse(results), nil
}

func (h *DefaultHandler) WriteSingleCoil(operation data.ModbusWriteSingleRequest[bool]) (response *data.WriteSingleCoilResponse, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.logger.Debug("WriteSingleCoil", zap.Uint16("Offset", operation.Offset()), zap.Bool("Value", operation.Value()))
	if int(operation.Offset()+1) > len(h.Coils) {
		return nil, common.ErrIllegalDataAddress
	}
	h.Coils[operation.Offset()+1] = operation.Value()
	return data.NewWriteSingleCoilResponse(operation.Offset(), operation.Value()), nil
}

func (h *DefaultHandler) WriteSingleRegister(operation data.ModbusWriteSingleRequest[uint16]) (response *data.WriteSingleRegisterResponse, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("WriteSingleRegister", zap.Uint16("Offset", operation.Offset()), zap.Uint16("Value", operation.Value()))
	if int(operation.Offset()+1) > len(h.HoldingRegisters) {
		return nil, common.ErrIllegalDataAddress
	}
	h.HoldingRegisters[operation.Offset()+1] = operation.Value()
	return data.NewWriteSingleRegisterResponse(operation.Offset(), operation.Value()), nil
}

func (h *DefaultHandler) WriteMultipleCoils(operation data.ModbusWriteArrayRequest[[]bool]) (response *data.WriteMultipleCoilsResponse, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("WriteMultipleCoils", zap.Uint16("Offset", operation.Offset()), zap.Bools("Values", operation.Values()))
	start, end := getRange(operation.Offset(), len(operation.Values()))
	if int(end) > len(h.Coils) {
		return nil, common.ErrIllegalDataAddress
	}
	for i, v := range operation.Values() {
		h.Coils[start+uint16(i)] = v
	}
	return data.NewWriteMultipleCoilsResponse(operation.Offset(), uint16(len(operation.Values()))), nil
}

func (h *DefaultHandler) WriteMultipleRegisters(operation data.ModbusWriteArrayRequest[[]uint16]) (response *data.WriteMultipleRegistersResponse, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.logger.Debug("WriteMultipleRegisters", zap.Uint16("Offset", operation.Offset()), zap.Uint16s("Values", operation.Values()))
	start, end := getRange(operation.Offset(), len(operation.Values()))
	if int(end) > len(h.HoldingRegisters) {
		return nil, common.ErrIllegalDataAddress
	}
	for i, v := range operation.Values() {
		h.HoldingRegisters[start+uint16(i)] = v
	}
	return data.NewWriteMultipleRegistersResponse(operation.Offset(), uint16(len(operation.Values()))), nil
}

func (h *DefaultHandler) Load(dataPath string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return err
	}
	coilsFilePath := filepath.Join(dataPath, coilsFile)
	discreteInputsFilePath := filepath.Join(dataPath, discreteInputsFile)
	holdingRegistersFilePath := filepath.Join(dataPath, holdingRegistersFile)
	inputRegistersFilePath := filepath.Join(dataPath, inputRegistersFile)
	if err := h.loadBoolArray(coilsFilePath, &h.Coils); err != nil {
		return err
	}
	if err := h.loadBoolArray(discreteInputsFilePath, &h.DiscreteInputs); err != nil {
		return err
	}
	if err := h.loadUInt16Array(holdingRegistersFilePath, &h.HoldingRegisters); err != nil {
		return err
	}
	if err := h.loadUInt16Array(inputRegistersFilePath, &h.InputRegisters); err != nil {
		return err
	}
	return nil
}

func (h *DefaultHandler) Save(dataPath string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dataPath, 0744); err != nil {
			return err
		}
	}
	coilsFilePath := filepath.Join(dataPath, coilsFile)
	discreteInputsFilePath := filepath.Join(dataPath, discreteInputsFile)
	holdingRegistersFilePath := filepath.Join(dataPath, holdingRegistersFile)
	inputRegistersFilePath := filepath.Join(dataPath, inputRegistersFile)
	if err := h.saveBoolArray(coilsFilePath, h.Coils); err != nil {
		return err
	}
	if err := h.saveBoolArray(discreteInputsFilePath, h.DiscreteInputs); err != nil {
		return err
	}
	if err := h.saveUInt16Array(holdingRegistersFilePath, h.HoldingRegisters); err != nil {
		return err
	}
	if err := h.saveUInt16Array(inputRegistersFilePath, h.InputRegisters); err != nil {
		return err
	}
	return nil
}

func (h *DefaultHandler) saveBoolArray(filename string, data []bool) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	return enc.Encode(data)
}

func (h *DefaultHandler) loadBoolArray(filename string, data *[]bool) error {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		h.logger.Warn("File not found, using defaults", zap.String("filename", filename))
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var fileData []bool
	err = decoder.Decode(&fileData)
	if err != nil {
		return err
	}
	arrLength := len(fileData)
	if arrLength > cap(*data) {
		*data = make([]bool, len(fileData))
	}
	copy(*data, fileData)
	return nil
}

func (h *DefaultHandler) saveUInt16Array(filename string, data []uint16) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	return enc.Encode(data)
}

func (h *DefaultHandler) loadUInt16Array(filename string, data *[]uint16) error {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		h.logger.Warn("File not found, using defaults", zap.String("filename", filename))
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var fileData []uint16
	err = decoder.Decode(&fileData)
	if err != nil {
		return err
	}
	arrLength := len(fileData)
	if arrLength > cap(*data) {
		*data = make([]uint16, len(fileData))
	}
	copy(*data, fileData)
	return nil
}
