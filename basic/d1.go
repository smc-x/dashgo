// Package basic provides fundamental methods for interacting with Dashgo.
package basic

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/smc-x/dashgo/internal/serial"
)

// D1 provides basic methods for interacting with Dashgo D1.
type D1 struct{}

// ValBaud gets the Baud rate of D1. The returned value should always be 115200. This method is
// useful for validating the connection with D1.
func (d1 *D1) ValBaud(dev serial.Device) (baud int, err error) {
	resp, err := dev.Request([]byte("b\r"))
	if err != nil {
		return
	}

	baudStr := string(bytes.TrimSpace(resp))
	baud, err = strconv.Atoi(baudStr)
	if err != nil {
		err = fmt.Errorf("cannot parse Baud rate %s: %v", baudStr, err)
	}
	return
}

// ValCounters gets the encoder counter values of D1. Note the counters accumulate the encoder
// values, with a range from -32768 to 32767.
func (d1 *D1) ValCounters(dev serial.Device) (leftAcc, rightAcc int, err error) {
	resp, err := dev.Request([]byte("e\r"))
	if err != nil {
		return
	}

	fields := strings.Fields(string(bytes.TrimSpace(resp)))
	if len(fields) != 2 {
		err = fmt.Errorf("expect 2 counter fields, but get %d", len(fields))
		return
	}

	values := make([]int, len(fields))
	for i, field := range fields {
		values[i], err = strconv.Atoi(field)
		if err != nil {
			err = fmt.Errorf("cannot parse counter field %s: %v", field, err)
			return
		}
	}

	return values[0], values[1], nil
}

// ValSonar gets the sonar values of D1. There are 4 sonar values in centimeters, collected from
// sensors armed at the front left, the front middle, the front right, and the back middle,
// respectively.
func (d1 *D1) ValSonar(dev serial.Device) (fL, fM, fR, bM int, err error) {
	resp, err := dev.Request([]byte("p\r"))
	if err != nil {
		return
	}

	fields := strings.Fields(string(bytes.TrimSpace(resp)))
	if len(fields) != 4 { // nolint:gomnd
		err = fmt.Errorf("expect 4 sonar fields, but get %d", len(fields))
		return
	}

	values := make([]int, len(fields))
	for i, field := range fields {
		values[i], err = strconv.Atoi(field)
		if err != nil {
			err = fmt.Errorf("cannot parse sonar field %s: %v", field, err)
			return
		}
	}

	return values[0], values[1], values[2], values[3], nil
}

// OpResetCounters resets the encoder counters.
func (d1 *D1) OpResetCounters(dev serial.Device) (text string, err error) {
	resp, err := dev.Request([]byte("r\r"))
	if err != nil {
		return
	}
	text = string(bytes.TrimSpace(resp))
	return
}

// OpSetEncoder sets the expected encoder values.
func (d1 *D1) OpSetEncoder(dev serial.Device, left, right int) error {
	return dev.Write([]byte(fmt.Sprintf("z %d %d\r", left, right)))
}

// OpSetPID sets the parameters for the internal PID controller.
func (d1 *D1) OpSetPID(dev serial.Device, kp, kd, ki, ko int) (text string, err error) {
	resp, err := dev.Request([]byte(fmt.Sprintf("u %d:%d:%d:%d\r", kp, kd, ki, ko)))
	if err != nil {
		return
	}
	text = string(bytes.TrimSpace(resp))
	return
}
