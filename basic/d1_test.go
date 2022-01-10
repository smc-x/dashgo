package basic_test

import (
	"testing"
	"time"

	"github.com/smc-x/dashgo/basic"
	"github.com/smc-x/dashgo/internal"
	"github.com/smc-x/dashgo/internal/serial"
)

// nolint:gocyclo
func TestD1(t *testing.T) {
	devices := internal.FindUSBDev([]string{"ttyUSB"})
	name, id := "", "1a86:7523"
	for name_, id_ := range devices {
		if id_ == id {
			name = name_
			break
		}
	}
	if name == "" {
		t.Fatal("Dashgo D1 not found")
	}

	d1 := &basic.D1{}
	err := serial.Session(name, 115200, func(dev serial.Device) error {
		baud, err := d1.ValBaud(dev)
		if err != nil {
			return err
		}
		t.Logf("Baud rate: %d", baud)

		text, err := d1.OpSetPID(dev, 50, 20, 0, 50)
		if err != nil {
			return err
		}
		t.Log(text)

		fL, fM, fR, bM, err := d1.ValSonar(dev)
		if err != nil {
			return err
		}
		t.Logf("Sonar values: %d, %d, %d, %d", fL, fM, fR, bM)
		if fL < 50 || fM < 50 || fR < 50 || bM < 50 {
			t.Fatal("Please spare enough room for testing Dashgo D1")
		}

		text, err = d1.OpResetCounters(dev)
		if err != nil {
			return err
		}
		t.Log(text)

		leftAcc, rightAcc, err := d1.ValCounters(dev)
		if err != nil {
			return err
		}
		t.Logf("Counters: %d, %d", leftAcc, rightAcc)

		err = d1.OpSetEncoder(dev, 5, -5)
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)

		err = d1.OpSetEncoder(dev, -5, 5)
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)

		leftAcc, rightAcc, err = d1.ValCounters(dev)
		if err != nil {
			return err
		}
		t.Logf("Counters (after rotation): %d, %d", leftAcc, rightAcc)

		text, err = d1.OpResetCounters(dev)
		if err != nil {
			return err
		}
		t.Log(text)

		leftAcc, rightAcc, err = d1.ValCounters(dev)
		if err != nil {
			return err
		}
		t.Logf("Counters: %d, %d", leftAcc, rightAcc)

		err = d1.OpSetEncoder(dev, 5, 5)
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)

		err = d1.OpSetEncoder(dev, -5, -5)
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)

		leftAcc, rightAcc, err = d1.ValCounters(dev)
		if err != nil {
			return err
		}
		t.Logf("Counters (after forth-back): %d, %d", leftAcc, rightAcc)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
