// Package serial wraps around low-level serial communications, attempting to make the interactions
// with devices through serial ports a bit more friendly.
package serial

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus" // nolint:depguard
	"github.com/tarm/serial"
)

// DefaultBufLen defines the default buffer length for reading from serial ports.
const DefaultBufLen = 128

var logSession = logrus.WithField("name", "Session")

// Session is a convenient context for performing serial communications.
func Session(name string, baud int, fn func(dev Device) error) error {
	return Session_(
		&serial.Config{
			Name:        name,
			Baud:        baud,
			ReadTimeout: 500 * time.Millisecond,
		},
		DefaultBufLen,
		true,
		fn,
	)
}

// Session_ is similar to Session but exports more details. The interface is subject to change.
func Session_(
	config *serial.Config,
	bufLen int,
	cleanOnStart bool,
	fn func(dev Device) error,
) error {
	port, err := serial.OpenPort(config)
	if err != nil {
		err = fmt.Errorf("failed opening serial port: %v", err)
		return err
	}
	defer func() {
		if errClose := port.Close(); errClose != nil {
			logSession.Errorf("failed closing the session: %v", errClose)
		}
	}()

	d := &device{bufLen: bufLen, port: port}
	if cleanOnStart {
		err = d.Clean()
		if err != nil {
			return err
		}
	}

	return fn(d)
}
