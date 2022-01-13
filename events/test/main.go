package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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
	Path2failures      = "./failures"
	NotConsecutiveFlag = 10
	SpeedBase          = 20
)

var (
	numFailures = 0
	speed_      = SpeedBase
	speedLock   = &sync.RWMutex{}
	startTime   = time.Now().Unix()
)

func speed() int {
	speedLock.RLock()
	defer speedLock.RUnlock()
	return speed_
}

func updateSpeed(s int) {
	speedLock.Lock()
	defer speedLock.Unlock()
	speed_ = s
}

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

			fields := strings.Split(msg, ",")
			if len(fields) != 2 {
				logMain.Errorf("expect 2 fields separated by comma, get %d", len(fields))
				return
			}
			ts, err := strconv.ParseFloat(fields[0], 64)
			if err != nil {
				logMain.Errorf("failed parsing timestamp: %v", err)
				return
			}
			if time.Now().UnixMilli()-int64(1000*ts) >= 1000 {
				logMain.Warn("drop stale instruction")
				return
			}

			fields = strings.Split(fields[1], ":")
			if len(fields) != 2 {
				logMain.Errorf("expect 2 fields separated by colon, get %d", len(fields))
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
					s := 2*SpeedBase - speed()%(2*SpeedBase)
					s /= 4
					errOp = d1.OpSetEncoder(dev, -s, s)
				case keys.AbsXRight:
					s := 2*SpeedBase - speed()%(2*SpeedBase)
					s /= 4
					errOp = d1.OpSetEncoder(dev, s, -s)
				}
			case keys.AbsY:
				switch value {
				case keys.AbsYUp:
					s := speed()
					errOp = d1.OpSetEncoder(dev, s, s)
				case keys.AbsYDown:
					s := speed()
					errOp = d1.OpSetEncoder(dev, -s, -s)
				}
			case keys.KeyX:
				if value == keys.KeyPush {
					errOp = d1.OpSetEncoder(dev, 0, 0)
				}
			case keys.KeyA:
				if value == keys.KeyPush {
					updateSpeed(SpeedBase)
				}
			case keys.KeyB:
				if value == keys.KeyPush {
					updateSpeed(2 * SpeedBase)
				}
			case keys.KeyY:
				if value == keys.KeyPush {
					updateSpeed(3 * SpeedBase)
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
