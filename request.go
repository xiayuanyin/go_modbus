package modbus

import (
	"encoding/binary"
	"fmt"
)

type ModbusRequest interface {
	toBytes() []byte
	ensureAddress(starsWithOne bool)
}

type ModbusReadRequest struct {
	TransId      uint16
	SlaveId      uint8
	FunctionCode FunctionCode
	Address      uint16
	FixedAddress uint16
	Length       uint16
	EnableCrc16  bool
}

func (req *ModbusReadRequest) ensureAddress(starsWithOne bool) {
	if starsWithOne {
		req.FixedAddress = req.Address + 1
	} else {
		req.FixedAddress = req.Address
	}
}
func (req *ModbusWriteRequest) ensureAddress(starsWithOne bool) {
	if starsWithOne {
		req.FixedAddress = req.Address + 1
	} else {
		req.FixedAddress = req.Address
	}
}
func (req *ModbusReadRequest) toBytes() []byte {
	totalBufferLength := 12
	packetLength := 6
	if req.EnableCrc16 {
		totalBufferLength += 2 // for crc16
		packetLength += 2      // for crc16
	}

	buf := make([]byte, totalBufferLength)
	binary.BigEndian.PutUint16(buf[0:], req.TransId)
	binary.BigEndian.PutUint16(buf[2:], 0)
	binary.BigEndian.PutUint16(buf[4:], uint16(packetLength))
	buf[6] = req.SlaveId
	buf[7] = byte(req.FunctionCode)
	binary.BigEndian.PutUint16(buf[8:], req.FixedAddress)
	binary.BigEndian.PutUint16(buf[10:], req.Length)

	if req.EnableCrc16 {
		crc := Crc16(buf[6 : len(buf)-2])
		binary.LittleEndian.PutUint16(buf[len(buf)-2:], crc)
	}

	fmt.Println("read request:", buf)
	return buf
}

func (req *ModbusWriteRequest) toBytes() []byte {
	isSingle := req.FunctionCode == FunctionCodeWriteSingleCoil || req.FunctionCode == FunctionCodeWriteSingleRegister
	bodyBytes := len(req.Buffer)
	// 7 = slaveId(1) + functionCode(1) + address(2) + count(2) + byteCount(1)
	packetLength := 7 + bodyBytes
	// 6 = transId(2) + protocolId(2) + packetLength(2)
	totalBufferLength := packetLength + 6
	if isSingle {
		bodyBytes = 2
		// coil or register的包会固定位8
		packetLength = 6
		totalBufferLength = 12
	}

	if req.EnableCrc16 {
		totalBufferLength += 2 // for crc16
		packetLength += 2      // for crc16
	}

	buf := make([]byte, totalBufferLength)
	binary.BigEndian.PutUint16(buf[0:], req.TransId)
	binary.BigEndian.PutUint16(buf[2:], 0)
	binary.BigEndian.PutUint16(buf[4:], uint16(packetLength))

	buf[6] = req.SlaveId
	buf[7] = byte(req.FunctionCode)
	binary.BigEndian.PutUint16(buf[8:], req.FixedAddress)
	if isSingle {
		copy(buf[10:], req.Buffer)
		//binary.BigEndian.PutUint16(buf[10:], binary.BigEndian.Uint16(req.Buffer))
	} else {
		binary.BigEndian.PutUint16(buf[10:], req.Count)
		buf[12] = byte(bodyBytes)
		copy(buf[13:], req.Buffer)
		//buf = append(buf[:13], req.Buffer...)
	}
	if req.EnableCrc16 {
		crc := Crc16(buf[6 : len(buf)-2])
		binary.LittleEndian.PutUint16(buf[len(buf)-2:], crc)
	}
	fmt.Println("write request:", buf)
	return buf
}

type ModbusWriteRequest struct {
	TransId      uint16
	SlaveId      uint8
	FunctionCode FunctionCode
	EnableCrc16  bool
	Address      uint16
	FixedAddress uint16
	Count        uint16 // coils or registers count
	Buffer       []byte
}
