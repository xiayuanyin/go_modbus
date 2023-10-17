package go_modbus

type ModbusClient interface {
	Connect() error
	Disconnect() error

	SendRequest(req ModbusRequest) error

	ReadCoils(address uint16, length uint16) (ResponseData[bool], error)
	ReadCoilsWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[bool], error)
	ReadDiscreteInputs(address uint16, length uint16) (ResponseData[bool], error)
	ReadDiscreteInputsWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[bool], error)
	ReadHoldingRegisters(address uint16, length uint16) (ResponseData[uint16], error)
	ReadHoldingRegistersWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[uint16], error)
	ReadInputRegisters(address uint16, length uint16) (ResponseData[uint16], error)
	ReadInputRegistersWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[uint16], error)

	WriteSingleCoil(address uint16, value bool) error
	WriteSingleCoilWithSlaveId(slaveId byte, address uint16, value bool) error
	WriteSingleRegister(address uint16, value uint16) error
	WriteSingleRegisterWithSlaveId(slaveId byte, address uint16, value uint16) error
	WriteCoils(address uint16, values []bool) error
	WriteCoilsWithSlaveId(slaveId byte, address uint16, values []bool) error
	WriteHoldingRegisters(address uint16, values []uint16) error
	WriteHoldingRegistersWithSlaveId(slaveId byte, address uint16, values []uint16) error
}
