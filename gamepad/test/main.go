package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus" // nolint:depguard
	"gopkg.in/yaml.v2"

	"github.com/smc-x/dashgo/basic"
	"github.com/smc-x/dashgo/internal"
	"github.com/smc-x/dashgo/internal/serial"
)

var logMain = logrus.WithField("name", "main")

const (
	Baud               = 115200
	Speed              = 10
	Path2failures      = "./failures"
	NotConsecutiveFlag = 10
)

var (
	numFailures = 0
	startTime   = time.Now().Unix()
)

// nolint:funlen,gocyclo
func main() {
	data, err := os.ReadFile(Path2failures)
	if err == nil {
		numFailures, _ = strconv.Atoi(string(bytes.TrimSpace(data)))
	}
	if numFailures >= 3 {
		time.Sleep(3 * time.Second)
	}

	defer func() {
		if r := recover(); r == nil {
			return
		}

		if time.Now().Unix()-startTime > NotConsecutiveFlag {
			numFailures = 1
		} else {
			numFailures++
		}

		_ = os.WriteFile(Path2failures, []byte(fmt.Sprintf("%d", numFailures)), os.ModePerm)
	}()

	// Load keys
	path2keys := "./keys.yaml"
	_, err = os.Stat(path2keys)
	if err != nil {
		path2keys = "../keys.yaml"
	}
	data, err = os.ReadFile(path2keys)
	if err != nil {
		logMain.Panicf("cannot read %s: %v", path2keys, err)
	}

	keys := &Keys{}
	if err = yaml.Unmarshal(data, keys); err != nil {
		logMain.Panicf("cannot unmarshal keys: %v", err)
	}
	logMain.Infof("keys: %v", keys)

	// Load config
	path2config := "./config.yaml"
	_, err = os.Stat(path2config)
	if err != nil {
		path2config = "../../config/config.yaml"
	}
	data, err = os.ReadFile(path2config)
	if err != nil {
		logMain.Panicf("cannot read %s: %v", path2config, err)
	}

	config := &Config{}
	if err = yaml.Unmarshal(data, config); err != nil {
		logMain.Panicf("cannot unmarshal config: %v", err)
	}
	logMain.Infof("config: %v", config)

	// Find Dashgo D1
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

	// Connect to NATS server
	nc, err := nats.Connect(config.URL)
	if err != nil {
		logMain.Panicf("cannot connect to NATS: %v", err)
	}
	defer nc.Close()

	d1 := &basic.D1{}
	err = serial.Session(name, Baud, func(dev serial.Device) error {
		_, errD1 := d1.ValBaud(dev)
		if errD1 != nil {
			return errD1
		}

		// Simple Async Subscriber
		_, errSub := nc.Subscribe(config.Key, func(m *nats.Msg) {
			msg := string(m.Data)
			if msg == config.Msg {
				return
			}
			logMain.Infof("received: %s", msg)

			fields := strings.Split(msg, ":")
			if len(fields) != 2 {
				logMain.Errorf("expect 2 fields, get %d", len(fields))
				return
			}

			code, errParse := strconv.Atoi(fields[0])
			if errParse != nil {
				logMain.Errorf("failed parsing the code: %v", err)
			}
			value, errParse := strconv.Atoi(fields[1])
			if errParse != nil {
				logMain.Errorf("failed parsing the value: %v", err)
			}

			var errOp error
			switch code {
			case keys.AbsX:
				switch value {
				case keys.AbsXLeft:
					errOp = d1.OpSetEncoder(dev, -Speed, Speed)
				case keys.AbsXRight:
					errOp = d1.OpSetEncoder(dev, Speed, -Speed)
				}
			case keys.AbsY:
				switch value {
				case keys.AbsYUp:
					errOp = d1.OpSetEncoder(dev, Speed, Speed)
				case keys.AbsYDown:
					errOp = d1.OpSetEncoder(dev, -Speed, -Speed)
				}
			case keys.KeyX:
				if value == keys.KeyPush {
					errOp = d1.OpSetEncoder(dev, 0, 0)
				}
			}
			if errOp != nil {
				logMain.Error(errOp)
			}
		})
		if errSub != nil {
			return fmt.Errorf("failed subscribing: %v", errSub)
		}

		notify := make(chan os.Signal, 1)
		signal.Notify(notify, syscall.SIGINT, syscall.SIGTERM)
		<-notify

		return nil
	})
	if err != nil {
		logMain.Panic(err)
	}
}
