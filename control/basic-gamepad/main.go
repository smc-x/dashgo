package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/smc-x/dashgo/basic"
	"github.com/smc-x/dashgo/internal"
	"github.com/smc-x/dashgo/internal/serial"
)

var logMain = logrus.WithField("name", "main")

func main() {
	defer func() {
		if e := recover(); e != nil {
			time.Sleep(time.Second)
			panic(e)
		}
	}()

	config := loadConfig([]string{
		"./config.yaml",
		"../../config/config.yaml",
	})
	ctrl := &control{}

	nc, err := nats.Connect(config.URL)
	if err != nil {
		logMain.Panicf("cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	devices := internal.FindUSBDev([]string{"ttyUSB"})
	name := ""
	for name_, id_ := range devices {
		if id_ == config.D1 {
			name = name_
			break
		}
	}
	if name == "" {
		logMain.Panic("Dashgo D1 not found")
	}

	d1 := &basic.D1{}
	err = serial.Session(name, config.Baud, func(dev serial.Device) error {
		_, err := d1.ValBaud(dev)
		if err != nil {
			logMain.Panic(err)
		}

		_, err = d1.OpResetCounters(dev)
		if err != nil {
			logMain.Panic(err)
		}

		_, err = d1.OpSetPID(dev, config.Kp, config.Kd, config.Ki, config.Ko)
		if err != nil {
			logMain.Panic(err)
		}

		_, err = nc.Subscribe("gamepad", func(m *nats.Msg) {
			msg := &Msg{}
			json.Unmarshal(m.Data, msg)

			ts := int(time.Now().UnixMilli())
			if ts-msg.Ts > 1000 {
				return
			}

			left, right := ctrl.update(ts, msg.Pl)
			errSet := d1.OpSetEncoder(dev, left, right)
			if errSet != nil {
				logMain.Panic(errSet)
			}
		})
		if err != nil {
			logMain.Panicf("failed subscribing: %v", err)
		}

		return nil
	})
	if err != nil {
		logMain.Panic(err)
	}

	notify := make(chan os.Signal, 1)
	signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM)
	<-notify
}
