package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCoilsResponse_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	response := ReadCoilsResponse{values: values}
	expected := []byte{0x02, 0x81, 0x81}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestReadCoilsRequest_Bytes(t *testing.T) {
	request := ReadCoilsRequest{offset: 0, count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestReadDiscreteInputsResponse_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	response := ReadDiscreteInputsResponse{values: values}
	expected := []byte{0x02, 0x81, 0x81}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestReadDiscreteInputsRequest_Bytes(t *testing.T) {
	request := ReadDiscreteInputsRequest{offset: 0, count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestReadHoldingRegistersResponse_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	response := ReadHoldingRegistersResponse{values: values}
	expected := []byte{0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestReadHoldingRegistersRequest_Bytes(t *testing.T) {
	request := ReadHoldingRegistersRequest{offset: 0, count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestReadInputRegistersResponse_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	response := ReadInputRegistersResponse{values: values}
	expected := []byte{0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestReadInputRegistersRequest_Bytes(t *testing.T) {
	request := ReadInputRegistersRequest{offset: 0, count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestWriteSingleCoilResponse_Bytes(t *testing.T) {
	response := WriteSingleCoilResponse{offset: 0, value: true}
	expected := []byte{0x00, 0x00, 0xFF, 0x00}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestWriteSingleCoilRequest_Bytes(t *testing.T) {
	request := WriteSingleCoilRequest{offset: 0, value: true}
	expected := []byte{0x00, 0x00, 0xFF, 0x00}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestWriteSingleRegisterResponse_Bytes(t *testing.T) {
	response := WriteSingleRegisterResponse{offset: 0, value: 0x0001}
	expected := []byte{0x00, 0x00, 0x00, 0x01}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestWriteSingleRegisterRequest_Bytes(t *testing.T) {
	request := WriteSingleRegisterRequest{offset: 0, value: 0x0001}
	expected := []byte{0x00, 0x00, 0x00, 0x01}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestWriteMultipleCoilsResponse_Bytes(t *testing.T) {
	response := WriteMultipleCoilsResponse{offset: 0, count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestWriteMultipleCoilsRequest_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	request := WriteMultipleCoilsRequest{offset: 0, values: values}
	expected := []byte{0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}

func TestWriteMultipleRegistersResponse_Bytes(t *testing.T) {
	response := WriteMultipleRegistersResponse{offset: 0, count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := ModbusOperationToBytes(response)
	assert.Equal(t, expected, result)
}

func TestWriteMultipleRegistersRequest_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	request := WriteMultipleRegistersRequest{offset: 0, values: values}
	expected := []byte{0x00, 0x00, 0x00, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := ModbusOperationToBytes(request)
	assert.Equal(t, expected, result)
}
