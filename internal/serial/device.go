package serial

import (
	"fmt"

	"github.com/tarm/serial"
)

// ErrTooManyData occurs when data written are less than expected.
var ErrTooManyData = fmt.Errorf("cannot write too many data at once")

// Device abstracts a serial device.
type Device interface {
	// Clean cleans dirty data, either not read or not transmitted.
	Clean() error
	// Read reads data from the device.
	Read() (data []byte, err error)
	// Write writes data to the device.
	Write(data []byte) error
	// Request writes an instruction to the device and then waits for a response.
	Request(ins []byte) (resp []byte, err error)
}

// device implements the Device interface.
type device struct {
	bufLen int
	port   *serial.Port
}

func (d *device) Clean() (err error) {
	err = d.port.Flush()
	if err != nil {
		err = fmt.Errorf("failed flushing dirty data: %v", err)
	}
	return
}

func (d *device) Read() (data []byte, err error) {
	// CAVEAT: buf is assumed large enough to suffice each single Read
	buf := make([]byte, d.bufLen)
	n, err := d.port.Read(buf)
	if err != nil {
		err = fmt.Errorf("failed reading from serial port: %v", err)
		return
	}
	data = buf[:n]
	return
}

func (d *device) Write(data []byte) error {
	// CAVEAT: data is assumed small enough to be written at once (or the dependent method can
	// handle large data automatically)
	n, err := d.port.Write(data)
	if err != nil {
		err = fmt.Errorf("failed writing to serial port: %v", err)
		return err
	}
	if n < len(data) {
		return ErrTooManyData
	}
	return nil
}

func (d *device) Request(ins []byte) (resp []byte, err error) {
	err = d.Write(ins)
	if err != nil {
		return
	}
	return d.Read()
}
