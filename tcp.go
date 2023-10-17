package go_modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type ResponseWrapper struct {
	ErrorCode    byte
	ErrorMessage string
	Data         []byte
}

type ResponseData[T bool | uint16] struct {
	Buffer []byte
	Data   []T
}

type ReadHandler struct {
	TransId uint16
	Chan    chan ResponseWrapper
}

type ModbusTcp struct {
	Host                string
	Port                int
	SlaveId             byte
	ConnectTimeout      time.Duration // TCP连接超时，默认10秒
	RequestTimeout      time.Duration // 请求超时，默认1秒
	AddressStartWithOne bool

	tcpStatus      TcpConnStatus
	socket         *net.Conn
	readHandlers   []ReadHandler
	pendingTcpData []byte
}

func (conn *ModbusTcp) Connect() error {
	if conn.ConnectTimeout == 0 {
		conn.ConnectTimeout = 10 * time.Second
	}
	if conn.RequestTimeout == 0 {
		conn.RequestTimeout = 1 * time.Second
	}
	if conn.SlaveId == 0 {
		conn.SlaveId = 1
	}
	socket, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", conn.Host, conn.Port), conn.ConnectTimeout)
	if err != nil {
		conn.tcpStatus = TcpDisconnected
		fmt.Println("err :", err)
		return err
	}
	conn.socket = &socket
	conn.tcpStatus = TcpConnected
	go func() {
		for {
			buf := make([]byte, 256)
			n, err := socket.Read(buf)
			if err == io.EOF {
				fmt.Println("socket closed")
				conn.tcpStatus = TcpDisconnected
				return
			} else if err != nil {
				fmt.Println("socket receive error:", err)
				conn.tcpStatus = TcpDisconnected
				return
			}
			conn.ReceiveData(buf[:n])
		}
	}()
	return nil
}
func (conn *ModbusTcp) IsConnected() bool {
	return conn.tcpStatus == TcpConnected
}
func (conn *ModbusTcp) Reopen() error {
	conn.tcpStatus = TcpConnecting
	return conn.Connect()
}

func (conn *ModbusTcp) Disconnect() error {
	if conn.socket == nil || conn.tcpStatus == TcpDisconnected {
		return nil
	}
	conn.tcpStatus = TcpDisconnected
	return (*conn.socket).Close()
}

func (conn *ModbusTcp) writeRequest(functionCode FunctionCode, slaveId byte, address uint16, value []byte, count uint16) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conn.RequestTimeout)
	defer cancel()
	transId := GetTransId()
	fmt.Println("value", value)
	var req ModbusRequest = &ModbusWriteRequest{
		TransId:      transId,
		SlaveId:      slaveId,
		FunctionCode: functionCode,
		Address:      address,
		Count:        count, // coils or registers count
		Buffer:       value,
	}
	err := conn.SendRequest(req)
	if err != nil {
		fmt.Println("!!!send write request error:", err)
		return nil, err
	}
	c := make(chan ResponseWrapper)
	handleMutex.Lock()
	conn.readHandlers = append(conn.readHandlers, ReadHandler{
		TransId: transId,
		Chan:    c,
	})
	handleMutex.Unlock()
	select {
	case <-ctx.Done():
		handleMutex.Lock()
		for i, _ := range conn.readHandlers {
			if conn.readHandlers[i].TransId == transId {
				//remove this handle
				close(conn.readHandlers[i].Chan)
				conn.readHandlers = append(conn.readHandlers[:i], conn.readHandlers[i+1:]...)
				break
			}
		}
		handleMutex.Unlock()
		return nil, fmt.Errorf("wait for transcation<%d> data timeout", transId)
	case res := <-c:
		if len(res.Data) == 0 && res.ErrorCode != 0 {
			var msg = ExceptionMessageMap[res.ErrorCode]
			if msg == "" {
				msg = "Unknown error"
			}
			return nil, fmt.Errorf("read holding registers error, error code: %d, %s", res.ErrorCode, msg)
		}
		return res.Data, nil
	}
}
func (conn *ModbusTcp) readRequest(functionCode FunctionCode, slaveId byte, address uint16, length uint16) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conn.RequestTimeout)
	defer cancel()
	transId := GetTransId()
	var req ModbusRequest = &ModbusReadRequest{
		TransId:      transId,
		SlaveId:      slaveId,
		FunctionCode: functionCode,
		Address:      address,
		Length:       length,
	}
	err := conn.SendRequest(req)
	if err != nil {
		fmt.Println("!!!send request error:", err)
		return nil, err
	}
	c := make(chan ResponseWrapper)
	handleMutex.Lock()
	conn.readHandlers = append(conn.readHandlers, ReadHandler{
		TransId: transId,
		Chan:    c,
	})
	handleMutex.Unlock()
	select {
	case <-ctx.Done():
		handleMutex.Lock()
		for i, _ := range conn.readHandlers {
			if conn.readHandlers[i].TransId == transId {
				//remove this handle
				close(conn.readHandlers[i].Chan)
				conn.readHandlers = append(conn.readHandlers[:i], conn.readHandlers[i+1:]...)
				break
			}
		}
		handleMutex.Unlock()
		return nil, fmt.Errorf("wait for transcation<%d> data timeout", transId)
	case res := <-c:
		if len(res.Data) == 0 && res.ErrorCode != 0 {
			var msg = ExceptionMessageMap[res.ErrorCode]
			if msg == "" {
				msg = "Unknown error"
			}
			return nil, fmt.Errorf("read holding registers error, error code: %d, %s", res.ErrorCode, msg)
		}
		return res.Data, nil
	}
}

func (conn *ModbusTcp) ReadHoldingRegistersWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[uint16], error) {
	var d ResponseData[uint16]
	buf, err := conn.readRequest(FunctionCodeReadHoldingRegisters, slaveId, address, length)
	if err == nil {
		d.Buffer = buf
		d.Data = ParseWord(buf)
	}
	return d, err
}
func (conn *ModbusTcp) ReadHoldingRegisters(address uint16, length uint16) (ResponseData[uint16], error) {
	return conn.ReadHoldingRegistersWithSlaveId(conn.SlaveId, address, length)
}

func (conn *ModbusTcp) ReadInputRegistersWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[uint16], error) {
	var d ResponseData[uint16]
	buf, err := conn.readRequest(FunctionCodeReadInputRegisters, slaveId, address, length)
	if err == nil {
		d.Buffer = buf
		d.Data = ParseWord(buf)
	}
	return d, err
}

func (conn *ModbusTcp) ReadInputRegisters(address uint16, length uint16) (ResponseData[uint16], error) {
	return conn.ReadInputRegistersWithSlaveId(conn.SlaveId, address, length)
}

func (conn *ModbusTcp) ReadCoilsWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[bool], error) {
	var d ResponseData[bool]
	buf, err := conn.readRequest(FunctionCodeReadCoils, slaveId, address, length)
	if err == nil {
		d.Buffer = buf
		d.Data = ParseCoils(buf, length)
	}
	return d, err
}

func (conn *ModbusTcp) ReadCoils(address uint16, length uint16) (ResponseData[bool], error) {
	return conn.ReadCoilsWithSlaveId(conn.SlaveId, address, length)
}

func (conn *ModbusTcp) ReadDiscreteInputsWithSlaveId(slaveId byte, address uint16, length uint16) (ResponseData[bool], error) {
	var d ResponseData[bool]
	buf, err := conn.readRequest(FunctionCodeReadDiscreteInputs, slaveId, address, length)
	if err == nil {
		d.Buffer = buf
		d.Data = ParseCoils(buf, length)
	}
	return d, err
}
func (conn *ModbusTcp) ReadDiscreteInputs(address uint16, length uint16) (ResponseData[bool], error) {
	return conn.ReadDiscreteInputsWithSlaveId(conn.SlaveId, address, length)
}

// write start
func (conn *ModbusTcp) WriteSingleCoilWithSlaveId(slaveId byte, address uint16, v bool) error {
	var value uint16
	if v {
		value = 0xFF00
	}
	_, err := conn.writeRequest(FunctionCodeWriteSingleCoil, slaveId, address, []byte{byte(value >> 8), byte(value)}, 1)
	return err
}
func (conn *ModbusTcp) WriteSingleCoil(address uint16, v bool) error {
	return conn.WriteSingleCoilWithSlaveId(conn.SlaveId, address, v)
}

func (conn *ModbusTcp) WriteSingleRegisterWithSlaveId(slaveId byte, address uint16, value uint16) error {
	_, err := conn.writeRequest(FunctionCodeWriteSingleRegister, slaveId, address, []byte{byte(value >> 8), byte(value)}, 1)
	return err
}
func (conn *ModbusTcp) WriteSingleRegister(address uint16, value uint16) error {
	return conn.WriteSingleRegisterWithSlaveId(conn.SlaveId, address, value)
}
func (conn *ModbusTcp) WriteCoilsWithSlaveId(slaveId byte, address uint16, values []bool) error {
	bitLength := len(values)
	//byteLength := bitLength / 8
	//if bitLength%8 > 0 {
	//	byteLength++
	//}
	buf := BoolToBytes(values)
	fmt.Println("write coils:", buf)
	_, err := conn.writeRequest(FunctionCodeWriteMultipleCoils, slaveId, address, buf, uint16(bitLength))
	return err
}
func (conn *ModbusTcp) WriteCoils(address uint16, values []bool) error {
	return conn.WriteCoilsWithSlaveId(conn.SlaveId, address, values)
}
func (conn *ModbusTcp) WriteHoldingRegistersWithSlaveId(slaveId byte, address uint16, values []uint16) error {
	buf := make([]byte, len(values)*2)
	for i := 0; i < len(values); i++ {
		binary.BigEndian.PutUint16(buf[i*2:], values[i])
	}
	_, err := conn.writeRequest(FunctionCodeWriteMultipleRegisters, slaveId, address, buf, uint16(len(values)))
	return err
}
func (conn *ModbusTcp) WriteHoldingRegisters(address uint16, values []uint16) error {
	return conn.WriteHoldingRegistersWithSlaveId(conn.SlaveId, address, values)
}

func (conn *ModbusTcp) SendRequest(req ModbusRequest) error {
	req.ensureAddress(conn.AddressStartWithOne)
	buf := req.toBytes()
	_, err := (*conn.socket).Write(buf)
	return err
}

var pendingMutex sync.Mutex
var handleMutex sync.Mutex

func (conn *ModbusTcp) ReceiveData(buf []byte) {
	if len(conn.pendingTcpData) > 0 {
		pendingMutex.Lock()
		buf = append(conn.pendingTcpData, buf...)
		//fmt.Println("追加pending数据:", conn.pendingTcpData)
		conn.pendingTcpData = []byte{}
		pendingMutex.Unlock()
	}
	var packetLength uint16 = 0
	if len(buf) > 9 {
		packetLength = binary.BigEndian.Uint16(buf[4:])
	}
	// TCP粘包
	if len(buf)-6 < int(packetLength) || int(packetLength) < 3 {
		pendingMutex.Lock()
		conn.pendingTcpData = append(conn.pendingTcpData, buf...)
		pendingMutex.Unlock()
		//fmt.Println("TCP粘包, pending data:", conn.pendingTcpData)
		return
	}
	//The MBAP header contains the following fields:
	// Transaction Identifier	2 Bytes	Identification of a MODBUS Request/ Response transaction.	Initialized by the client	Recopied by the server from the received request
	// Protocol Identifier	2 Bytes	0 = MODBUS protocol	Initialized by the client	Recopied by the server from the received request
	// PacketLength	2 Bytes	Number of following bytes	Initialized by the client (request)	Initialized by the server (Response)

	// Slave Identifier	1 Byte	Identification of a remote slave connected on a serial line or on other buses.	Initialized by the client	Recopied by the server from the received request
	// Function Code 1Byte
	// Exception Code 1Byte | ByteCount 1 Byte
	// Data N Bytes
	trnasId := binary.BigEndian.Uint16(buf[0:])
	protocolId := binary.BigEndian.Uint16(buf[2:])
	exceptionOrCount := buf[8]
	if protocolId != 0 {
		fmt.Printf("protocolId != 0, unkown data!, %v", buf)
		return
	}

	//slaveId := buf[6]
	//request Func  --> response Error Func Code
	//01 (01 hex) --> 129 (81 hex)
	//02 (02 hex) --> 130 (82 hex)
	//03 (03 hex) --> 131 (83 hex)
	//04 (04 hex) --> 132 (84 hex)
	//05 (05 hex) --> 133 (85 hex)
	//06 (06 hex) --> 134 (86 hex)
	//15 (0F hex) --> 143 (8F hex)
	//16 (10 hex) --> 144 (90 hex)
	//17 (11 hex) --> 145 (91 hex)
	//20 (14 hex) --> 148 (94 hex)
	//43 (2B hex) --> 171 (AB hex)
	//funcCode := buf[7]
	errCode := buf[8]
	//length包含3个字节信息， slaveId, FunctionCode, ByteCount
	fmt.Println("packetLength:", packetLength, "buf length:", len(buf), "response", buf)
	data := buf[9 : 6+packetLength]
	remainingData := buf[6+packetLength:]
	if len(remainingData) > 0 {
		conn.ReceiveData(remainingData)
	}
	//fmt.Println("receive data transId:", trnasId, "protocolId:", protocolId, "length:", length, "slaveId:", slaveId, "funcCode:", funcCode, "errCode:", errCode, "data:", buf)

	for i, _ := range conn.readHandlers {
		if conn.readHandlers[i].TransId == trnasId {
			response := ResponseWrapper{
				ErrorCode: errCode,
				Data:      data,
			}
			if exceptionOrCount > 0 && len(data) == 0 {
				msg := ExceptionMessageMap[exceptionOrCount]
				if msg == "" {
					response.ErrorMessage = "Unknown exception"
				}
			}
			conn.readHandlers[i].Chan <- response
			close(conn.readHandlers[i].Chan)
			// remove this handle
			handleMutex.Lock()
			conn.readHandlers = append(conn.readHandlers[:i], conn.readHandlers[i+1:]...)
			handleMutex.Unlock()
			break
		}
	}

}
