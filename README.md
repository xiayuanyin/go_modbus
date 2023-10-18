# go modbus
golang modbus tcp/rtu client library

-----

Allow to read and write coils, discrete inputs, holding registers and input registers from a Modbus server.

Currently only Modbus TCP is supported.

Allow asynchronous reading and writing of multiple registers/coils (TCP and if only slave server supported).

# Usage
```go
// local test server
var c ModbusClient = &(ModbusPort{
    Host:           "127.0.0.1",
    Port:           1502,
    ConnectTimeout: time.Second * 10,
    SlaveId:        1,
})
_ = c.Connect()
defer c.Disconnect()
var received int32 = 0
var success int32 = 0
wg := sync.WaitGroup{}
for i := 0; i < 1000; i++ {
	wg.Add(1)
	go func(ii int) {
		r, err := c.ReadHoldingRegisters(uint16(ii), 10)
		atomic.AddInt32(&received, 1)
		if err != nil {
			fmt.Println("err: ", err)
		} else {
			atomic.AddInt32(&success, 1)
			fmt.Println("read: \n", ii, "-->", r.Data)
		}
		wg.Done()
	}(i)
}
wg.Wait()
fmt.Println("received total: ", received, "success: ", success)

```

Supported functions
-------------------
Bit access:
*   Read Discrete Inputs
*   Read Coils
*   Write Single Coil
*   Write Multiple Coils

16-bit access:
*   Read Input Registers
*   Read Holding Registers
*   Write Single Register
*   Write Multiple Registers
*   Read/Write Multiple Registers


# TODO
- [x] modbus tcp client
- [x] modbus udp client
- [ ] modbus rtu client
