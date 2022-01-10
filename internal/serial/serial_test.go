package serial_test

import (
	"testing"

	"github.com/smc-x/dashgo/internal/serial"
)

func TestSession(t *testing.T) {
	err := serial.Session("/dev/ttyUSB0", 115200, func(dev serial.Device) error {
		resp, err := dev.Request([]byte("b\r"))
		if err != nil {
			return err
		}
		t.Logf("%s", string(resp))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
