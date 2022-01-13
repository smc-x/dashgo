package main

// Keys contains a set of prescribed gamepad codes.
type Keys struct {
	AbsX      int `yaml:"abs_x"`
	AbsXRight int `yaml:"abs_x_right"`
	AbsXLeft  int `yaml:"abs_x_left"`

	AbsY     int `yaml:"abs_y"`
	AbsYUp   int `yaml:"abs_y_up"`
	AbsYDown int `yaml:"abs_y_down"`

	KeyX    int `yaml:"key_x"`
	KeyY    int `yaml:"key_y"`
	KeyA    int `yaml:"key_a"`
	KeyB    int `yaml:"key_b"`
	KeyPush int `yaml:"key_push"`
}