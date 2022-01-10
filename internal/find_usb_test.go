package internal_test

import (
	"testing"

	"github.com/smc-x/dashgo/internal"
)

func TestFindUSBDev(t *testing.T) {
	devices := internal.FindUSBDev([]string{"ttyUSB", "ttyACM", "video"})
	for k, v := range devices {
		t.Log(k, v)
	}
}
