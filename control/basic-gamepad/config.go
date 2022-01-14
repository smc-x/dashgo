package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config encapsulates required parameters.
type Config struct {
	D1   string `yaml:"d1_id"`
	Baud int    `yaml:"d1_baud"`

	Kp int `yaml:"pid_kp"`
	Kd int `yaml:"pid_kd"`
	Ki int `yaml:"pid_ki"`
	Ko int `yaml:"pid_ko"`

	URL string `yaml:"nats_url"`
}

// loadConfig loads parameters from a list of path candidates.
func loadConfig(paths []string) *Config {
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			logMain.Debugf("cannot read %s: %v", path, err)
			continue
		}

		config := &Config{}
		if err = yaml.Unmarshal(data, config); err != nil {
			logMain.Debugf("cannot unmarshal config %s: %v", path, err)
			continue
		}
		return config
	}

	logMain.Panic("failed loading config")
	return nil
}
