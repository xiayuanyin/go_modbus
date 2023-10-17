package modbus

import (
	"encoding/binary"
	"sync"
)

func Crc16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) == 1 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func ParseWord(data []byte) []uint16 {
	length := len(data) / 2
	contents := make([]uint16, length)
	for i := 0; i < length; i++ {
		reg := binary.BigEndian.Uint16(data[i*2:])
		contents[i] = reg
	}
	return contents
}
func ParseCoils(data []byte, coilsLength uint16) []bool {
	contents := make([]bool, coilsLength)
	for i := 0; i < int(coilsLength); i++ {
		contents[i] = data[i/8]&(1<<uint(i%8)) > 0
	}
	return contents
}

var TransId uint16 = 0

var transMutex sync.Mutex

func GetTransId() uint16 {
	transMutex.Lock()
	TransId++
	if TransId == 0 {
		TransId++
	}
	transMutex.Unlock()
	return TransId
}

func BoolToBytes(data []bool) []byte {
	length := len(data)
	byteLength := length / 8
	if length%8 > 0 {
		byteLength++
	}
	contents := make([]byte, byteLength)
	for i := 0; i < length; i++ {
		if data[i] {
			contents[i/8] |= 1 << uint(i%8)
		}
	}
	return contents
}

func BytesToBool(data []byte) []bool {
	length := len(data) * 8
	contents := make([]bool, length)
	for i := 0; i < length; i++ {
		contents[i] = data[i/8]&(1<<uint(i%8)) > 0
	}
	return contents
}
