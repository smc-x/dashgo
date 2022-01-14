package main

import (
	"math"
)

// Gamepad abstracts the gamepad model.
type Gamepad struct {
	DirX []int `json:"dir_x"`
	DirY []int `json:"dir_y"`
	BtnX []int `json:"btn_x"`
	BtnY []int `json:"btn_y"`
	BtnA []int `json:"btn_a"`
	BtnB []int `json:"btn_b"`
	BtnS []int `json:"btn_s"`
}

// Msg defines the message struct.
type Msg struct {
	Ts int      `json:"ts"`
	Pl *Gamepad `json:"pl"`
}

// interpret extracts values from gamepad events.
func interpret(ts int, pair []int) (elapsed, value int) {
	if len(pair) != 2 {
		return
	}
	return ts - pair[0], pair[1]
}

// ratios = [2 * sigmoid(x) for x = -3.6, -3.2, ..., -0.4, 0.0]
var ratios = [10]float64{
	0.05319398715373171,
	0.07833144559352871,
	0.11464835179773751,
	0.16634539298784476,
	0.2384058440442351,
	0.33596322973215115,
	0.46295043300196476,
	0.6200510377447751,
	0.802624679775096,
	1.0,
}

// speedLevel defines a sequence of speed levels.
type speedLevel int

const (
	speedL0 speedLevel = iota
	speedL1
	speedL2
	speedL3
	speedN1
	speedN2
	speedN3
)

const (
	speedL0Val = 0.
	speedL1Val = 10.
	speedL2Val = 20.
	speedL3Val = 30.
)

// levelToVal converts speed levels to concrete values.
func levelToVal(le speedLevel) float64 {
	switch le {
	case speedL0:
		return speedL0Val
	case speedL1:
		return speedL1Val
	case speedL2:
		return speedL2Val
	case speedL3:
		return speedL3Val
	case speedN1:
		return -speedL1Val
	case speedN2:
		return -speedL2Val
	case speedN3:
		return -speedL3Val
	default:
		return speedL0Val
	}
}

// control encapsulates the control context.
type control struct {
	current float64

	gap   float64
	added float64

	expect speedLevel
}

// update calculates the speed to apply.
func (ctrl *control) update(ts int, gp *Gamepad) (int, int) {
	yLeft, yRight, stopped := ctrl.getDirY(ts, gp)
	if stopped {
		return int(yLeft), int(yRight)
	}

	xLeft, xRight := getDirX(ts, gp)

	yAbs := math.Abs(yLeft)
	xAbs := math.Abs(xLeft)
	if (xAbs + yAbs) < 1e-3 {
		return 0, 0
	}

	scale := 1.
	if xAbs > yAbs {
		scale = xAbs / (yAbs + xAbs)
	} else {
		scale = yAbs / (yAbs + xAbs)
	}

	return int(scale * (yLeft + xLeft)), int(scale * (yRight + xRight))
}

// getDirX gets the expected speed at the X direction (stateless).
func getDirX(ts int, gp *Gamepad) (left, right float64) {
	elapsedX, x := interpret(ts, gp.DirX)
	ind := elapsedX / 100
	if ind >= 10 {
		ind = 9
	}
	speed := speedL1Val * ratios[ind] / 2
	switch {
	case x < 0:
		return -speed, speed
	case x > 0:
		return speed, -speed
	default:
		return 0, 0
	}
}

// getDirY gets the expected speed at the Y direction (stateful).
func (ctrl *control) getDirY(ts int, gp *Gamepad) (left, right float64, stopped bool) {
	_, l0 := interpret(ts, gp.BtnA)
	elapsedL1, l1 := interpret(ts, gp.BtnB)
	elapsedL2, l2 := interpret(ts, gp.BtnX)
	elapsedL3, l3 := interpret(ts, gp.BtnY)
	elapsedY, y := interpret(ts, gp.DirY)

	// Decide the expected speed
	expect := speedL0
	elapsed := elapsedY
	switch {
	case l0 > 0:
		// Stop immediately
		ctrl.current = 0
		ctrl.gap = 0
		ctrl.added = 0
		ctrl.expect = speedL0
		return 0, 0, true
	case y == 0:
		// Stop smoothly
		expect = speedL0
	case l1 > 0:
		if elapsed > elapsedL1 {
			elapsed = elapsedL1
		}
		if y > 0 {
			expect = speedL1
		} else {
			expect = speedN1
		}
	case l2 > 0:
		if elapsed > elapsedL2 {
			elapsed = elapsedL2
		}
		if y > 0 {
			expect = speedL2
		} else {
			expect = speedN2
		}
	case l3 > 0:
		if elapsed > elapsedL3 {
			elapsed = elapsedL3
		}
		if y > 0 {
			expect = speedL3
		} else {
			expect = speedN3
		}
	default:
		if elapsed > elapsedL1 {
			elapsed = elapsedL1
		}
		if elapsed > elapsedL2 {
			elapsed = elapsedL2
		}
		if elapsed > elapsedL3 {
			elapsed = elapsedL3
		}
	}

	// Make correction
	if ctrl.expect != expect {
		ctrl.expect = expect
		ctrl.gap = levelToVal(expect) - ctrl.current
		ctrl.added = 0
	}
	if math.Abs(ctrl.gap) > 1e-3 {
		ctrl.current -= ctrl.added
		ind := elapsed / 100
		if ind >= 10 {
			ind = 9
		}
		ctrl.added = ratios[ind] * ctrl.gap
		ctrl.current += ctrl.added
		if ind == 9 {
			// Clear gap
			ctrl.gap = 0
			ctrl.added = 0
		}
	}

	return ctrl.current, ctrl.current, false
}
