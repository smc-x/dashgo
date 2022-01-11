package main

// Config contains a set of prescribed configuration values.
type Config struct {
	URL string `yaml:"nats_url"`
	D1  string `yaml:"d1_id"`
	Key string `yaml:"publisher_key"`
	Msg string `yaml:"alive_msg"`
}
