package modbus

type FunctionCode byte
type ExceptionCode byte
type TcpConnStatus int

const (
	FunctionCodeReadCoils                  FunctionCode = 0x01
	FunctionCodeReadDiscreteInputs         FunctionCode = 0x02
	FunctionCodeReadHoldingRegisters       FunctionCode = 0x03
	FunctionCodeReadInputRegisters         FunctionCode = 0x04
	FunctionCodeWriteSingleCoil            FunctionCode = 0x05
	FunctionCodeWriteSingleRegister        FunctionCode = 0x06
	FunctionCodeWriteMultipleCoils         FunctionCode = 0x0F
	FunctionCodeWriteMultipleRegisters     FunctionCode = 0x10
	FunctionCodeMaskWriteRegister          FunctionCode = 0x16
	FunctionCodeReadWriteMultipleRegisters FunctionCode = 0x17
)

var ExceptionMessageMap = map[byte]string{
	1:  "Illegal function",
	2:  "Illegal data address",
	3:  "Illegal data value",
	4:  "Slave device failure",
	5:  "Acknowledge",
	6:  "Slave device busy",
	7:  "Negative acknowledge",
	8:  "Memory parity error",
	10: "Gateway path unavailable",
	11: "Gateway target device failed to respond",
}

const (
	TcpConnecting   TcpConnStatus = 0
	TcpConnected    TcpConnStatus = 1
	TcpDisconnected TcpConnStatus = 2
)
