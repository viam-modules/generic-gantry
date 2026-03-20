package testrig

import (
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/resource"
)

// TestRig holds simulated components and configuration for integration testing.
type TestRig struct {
	Motor           *SimulatedMotor
	Board           *SimulatedBoard // nil for encoder-only
	MotorName       string
	BoardName       string
	LimitSwitchPins []string
	LimitPinEnabled bool
	LengthMm        float64
	MmPerRevolution float64
	GantryMmPerSec  float64
}

// Dependencies returns a resource.Dependencies map suitable for passing to newSingleAxis.
func (r *TestRig) Dependencies() resource.Dependencies {
	deps := make(resource.Dependencies)

	deps[motor.Named(r.MotorName)] = r.Motor
	if r.Board != nil {
		deps[board.Named(r.BoardName)] = r.Board
	}

	return deps
}

// NewPrinterXAxis creates a test rig simulating a desktop 3D printer belt-driven X axis
// with 2 limit switches, 300mm travel, 40mm per revolution.
func NewPrinterXAxis() *TestRig {
	m := NewSimulatedMotor("motor1", 4.0) // 160mm into 300mm range
	pins := map[string]*LimitSwitchPin{
		"pin0": NewLimitSwitchPin(m, 0.0, false, true), // min endstop
		"pin1": NewLimitSwitchPin(m, 7.5, true, true),  // max endstop at 300mm / 40mm/rev
	}
	b := NewSimulatedBoard("board1", pins)

	return &TestRig{
		Motor:           m,
		Board:           b,
		MotorName:       "motor1",
		BoardName:       "board1",
		LimitSwitchPins: []string{"pin0", "pin1"},
		LimitPinEnabled: true,
		LengthMm:        300,
		MmPerRevolution: 40,
		GantryMmPerSec:  500,
	}
}

// NewCNCMillXAxis creates a test rig simulating a CNC mill X axis
// with 2 limit switches, 600mm travel, 5mm per revolution (ball screw).
func NewCNCMillXAxis() *TestRig {
	m := NewSimulatedMotor("motor1", 60.0) // 300mm center
	pins := map[string]*LimitSwitchPin{
		"pin0": NewLimitSwitchPin(m, 0.0, false, true),  // min endstop
		"pin1": NewLimitSwitchPin(m, 120.0, true, true), // max endstop at 600mm / 5mm/rev
	}
	b := NewSimulatedBoard("board1", pins)

	return &TestRig{
		Motor:           m,
		Board:           b,
		MotorName:       "motor1",
		BoardName:       "board1",
		LimitSwitchPins: []string{"pin0", "pin1"},
		LimitPinEnabled: true,
		LengthMm:        600,
		MmPerRevolution: 5,
		GantryMmPerSec:  166,
	}
}

// NewLinearActuator creates a test rig simulating an encoder-only linear actuator
// with no limit switches, 500mm travel, 8mm per revolution.
func NewLinearActuator() *TestRig {
	m := NewSimulatedMotor("motor1", 0.0)

	return &TestRig{
		Motor:           m,
		Board:           nil,
		MotorName:       "motor1",
		BoardName:       "",
		LimitSwitchPins: nil,
		LimitPinEnabled: false,
		LengthMm:        500,
		MmPerRevolution: 8,
		GantryMmPerSec:  100,
	}
}

// NewSingleSwitchAxis creates a test rig with 1 limit switch,
// 1000mm travel, 10mm per revolution.
func NewSingleSwitchAxis() *TestRig {
	m := NewSimulatedMotor("motor1", 50.0) // 500mm mid-range
	pins := map[string]*LimitSwitchPin{
		"pin0": NewLimitSwitchPin(m, 0.0, false, true), // min endstop
	}
	b := NewSimulatedBoard("board1", pins)

	return &TestRig{
		Motor:           m,
		Board:           b,
		MotorName:       "motor1",
		BoardName:       "board1",
		LimitSwitchPins: []string{"pin0"},
		LimitPinEnabled: true,
		LengthMm:        1000,
		MmPerRevolution: 10,
		GantryMmPerSec:  200,
	}
}
