package main

import (
	"math"
)

// Gamepad abstracts the gamepad model.
type Gamepad struct {
	DirX []int64 `json:"dir_x"`
	DirY []int64 `json:"dir_y"`
	BtnX []int64 `json:"btn_x"`
	BtnY []int64 `json:"btn_y"`
	BtnA []int64 `json:"btn_a"`
	BtnB []int64 `json:"btn_b"`
	BtnS []int64 `json:"btn_s"`
}

// Msg defines the message struct.
type Msg struct {
	TS int64    `json:"ts"`
	Pl *Gamepad `json:"pl"`
}

// interpret extracts values from gamepad events.
func interpret(ts int64, pair []int64) (elapsed, value int64) {
	if len(pair) != 2 {
		return
	}
	return ts - pair[0], pair[1]
}

const (
	eps       = 1e-3
	lenRatios = 20
)

// ratios = [sigmoid(x) for x = -1.8, -1.6, ..., 1.8] + [1.0]
var ratios = [lenRatios]float64{
	0.14185106490048777,
	0.16798161486607552,
	0.19781611144141825,
	0.23147521650098238,
	0.2689414213699951,
	0.31002551887238755,
	0.3543436937742046,
	0.401312339887548,
	0.45016600268752216,
	0.5,
	0.5498339973124778,
	0.5986876601124521,
	0.6456563062257955,
	0.6899744811276125,
	0.7310585786300049,
	0.7685247834990175,
	0.8021838885585818,
	0.8320183851339246,
	0.8581489350995123,
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
	speedL1Val = 20.
	speedL2Val = 40.
	speedL3Val = 50.
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
func (ctrl *control) update(ts int64, gp *Gamepad) (left, right int) {
	yLeft, yRight, stopped := ctrl.getDirY(ts, gp)
	if stopped {
		return int(yLeft), int(yRight)
	}

	xLeft, xRight := getDirX(ts, gp)

	yAbs := math.Abs(yLeft)
	xAbs := math.Abs(xLeft)
	if (yAbs + xAbs) < eps {
		return 0, 0
	}

	var scale float64
	if xAbs > yAbs {
		scale = xAbs / (yAbs + xAbs)
	} else {
		scale = yAbs / (yAbs + xAbs)
	}

	return int(scale*(yLeft+xLeft) + 0.5), int(scale*(yRight+xRight) + 0.5)
}

// getDirX gets the expected speed at the X direction (stateless).
func getDirX(ts int64, gp *Gamepad) (left, right float64) {
	elapsedX, x := interpret(ts, gp.DirX)
	ind := elapsedX / 100
	if ind >= lenRatios {
		ind = lenRatios - 1
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
//
// nolint:funlen,gocyclo
func (ctrl *control) getDirY(ts int64, gp *Gamepad) (left, right float64, stopped bool) {
	_, l0 := interpret(ts, gp.BtnA)
	elapsedL1, l1 := interpret(ts, gp.BtnB)
	elapsedL2, l2 := interpret(ts, gp.BtnY)
	elapsedL3, l3 := interpret(ts, gp.BtnX)
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
	if math.Abs(ctrl.gap) > eps {
		ctrl.current -= ctrl.added
		ind := elapsed / 100
		if ind >= lenRatios {
			ind = lenRatios - 1
		}
		ctrl.added = ratios[ind] * ctrl.gap
		ctrl.current += ctrl.added
		if ind == lenRatios-1 {
			// Clear gap
			ctrl.gap = 0
			ctrl.added = 0
		}
	}

	return ctrl.current, ctrl.current, false
}
