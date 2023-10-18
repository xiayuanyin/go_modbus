package test

import (
	"fmt"
	"github.com/xiayuanyin/go_modbus"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCrc16(t *testing.T) {
	//01 03 61 00 00 02
	data := []byte{0x01, 0x03, 0x61, 0x00, 0x00, 0x02}
	crc := go_modbus.Crc16(data)
	if crc != 0xf7db {
		t.Error("crc16 error", crc)
	}
}

func TestTCPClientRead(t *testing.T) {
	var c go_modbus.ModbusClient = &(go_modbus.ModbusPort{
		Host:           "127.0.0.1",
		Port:           502,
		ConnectTimeout: time.Second * 5,
		RequestTimeout: time.Second * 1,
	})
	err := c.Connect()
	defer c.Disconnect()

	if err != nil {
		t.Error("connect error", err)
		return
	}
	d, _ := c.ReadHoldingRegisters(0, 10)
	if len(d.Data) != 10 {
		t.Error("read error 10 holding registers error", d.Buffer)
	}
	err = c.WriteHoldingRegisters(uint16(0), []uint16{uint16(1), uint16(99), uint16(123)})
	if err != nil {
		t.Error("write error", err)
	}
	d, _ = c.ReadHoldingRegisters(0, 10)
	if len(d.Data) != 10 {
		t.Error("read error 10 holding registers error", d.Buffer)
	}
	if d.Data[0] != 1 || d.Data[1] != 99 || d.Data[2] != 123 {
		t.Error("write registers failed, now read data", d.Data)
	}
	err = c.WriteSingleRegister(uint16(8), 888)
	if err != nil {
		t.Error("write single register error", err)
	}
	//return
	err = c.WriteCoils(uint16(16), []bool{true, false, true, false, true, false, true, false})
	if err != nil {
		t.Error("write coils error", err)
	}

	coils, _ := c.ReadCoils(16, 8)
	if len(coils.Data) != 8 {
		t.Error("read coils error", coils.Buffer)
	} else if coils.Data[0] != true || coils.Data[1] != false || coils.Data[2] != true || coils.Data[3] != false ||
		coils.Data[4] != true || coils.Data[5] != false || coils.Data[6] != true || coils.Data[7] != false {
		t.Error("write coils failed, now read data", coils.Data)
	}

	err = c.WriteSingleCoil(uint16(32), true)
	if err != nil {
		t.Error("write single coil error", err)
	}

	coils, _ = c.ReadCoils(32, 1)
	if len(coils.Data) != 1 {
		t.Error("read coils error", coils.Buffer)
	} else if coils.Data[0] != true {
		t.Error("write coils failed, now read data", coils.Data)
	}

	var received int32 = 0
	var success int32 = 0
	wg := sync.WaitGroup{}
	// 并行读取100次
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(ii int) {
			_, err := c.ReadHoldingRegisters(uint16(ii/10), 1)
			atomic.AddInt32(&received, 1)
			if err != nil {
				fmt.Println("ReadHoldingRegisters err: ", err)
			} else {
				atomic.AddInt32(&success, 1)
				//fmt.Println("read: \n", ii, "-->", r.Data)
			}
			wg.Done()
		}(i)

	}
	wg.Wait()
	if received != 100 {
		t.Error("received count error", received)
	}
	if success != 100 {
		t.Error("success count error", success)
	}

	startTime := time.Now()
	for i := 0; i < 10; i++ {
		err = c.WriteHoldingRegisters(uint16(i), []uint16{uint16(i)})
	}
	fmt.Println("write 10 registers cost: ", time.Since(startTime))

	err = c.WriteSingleRegister(6, 255)
	if err != nil {
		t.Error("write single register error", err)
	}
	res, err := c.ReadHoldingRegisters(6, 1)
	if err != nil || res.Data[0] != 255 {
		t.Error("write single register error", res.Data)
	}

	err = c.WriteSingleCoil(11, true)
	if err != nil {
		t.Error("write single coil error", err)
	}
	res2, _ := c.ReadCoils(11, 8)
	if len(res2.Data) != 8 || res2.Data[0] != true {
		t.Error("write single coil error", res2.Data)
	}
}
