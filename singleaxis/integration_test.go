package singleaxis

import (
	"context"
	"math"
	"testing"

	"go.viam.com/rdk/components/gantry"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/test"

	"github.com/viam-modules/generic-gantry/testrig"
)

func rigConfig(name string, rig *testrig.TestRig) resource.Config {
	cfg := &Config{
		Motor:           rig.MotorName,
		LengthMm:        rig.LengthMm,
		MmPerRevolution: rig.MmPerRevolution,
		GantryMmPerSec:  rig.GantryMmPerSec,
	}
	if len(rig.LimitSwitchPins) > 0 {
		cfg.Board = rig.BoardName
		cfg.LimitSwitchPins = rig.LimitSwitchPins
		cfg.LimitPinEnabled = &rig.LimitPinEnabled
	}

	return resource.Config{
		Name:                name,
		ConvertedAttributes: cfg,
	}
}

func TestIntegrationHomeTwoSwitch(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewPrinterXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("printer-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	homed, err := g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homed, test.ShouldBeTrue)

	sa := g.(*singleAxis)
	test.That(t, sa.positionLimits[0], test.ShouldBeBetween, -1.0, 1.0)
	test.That(t, sa.positionLimits[1], test.ShouldBeBetween, 6.5, 8.5)
	test.That(t, sa.positionRange, test.ShouldBeGreaterThan, 0)

	pos, err := g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, len(pos), test.ShouldEqual, 1)
	test.That(t, pos[0], test.ShouldBeBetween, 120, 180) // center ~150mm
}

func TestIntegrationHomeEncoderOnly(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewLinearActuator()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("actuator", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	homed, err := g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homed, test.ShouldBeTrue)

	sa := g.(*singleAxis)
	// 500mm / 8mm per rev = 62.5 revolutions
	test.That(t, sa.positionLimits[0], test.ShouldEqual, 0.0)
	test.That(t, sa.positionLimits[1], test.ShouldEqual, 62.5)
	test.That(t, sa.positionRange, test.ShouldEqual, 62.5)
}

func TestIntegrationHomeOneSwitch(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewSingleSwitchAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("single-sw", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	homed, err := g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homed, test.ShouldBeTrue)

	sa := g.(*singleAxis)
	test.That(t, sa.positionLimits[0], test.ShouldBeBetween, -1.0, 1.0)
	// positionB = positionA + 1000/10 = positionA + 100
	test.That(t, sa.positionLimits[1], test.ShouldBeBetween, 99.0, 101.0)
}

func TestIntegrationMoveAfterHoming(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewPrinterXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("printer-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	_, err = g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)

	// Move to 100mm
	err = g.MoveToPosition(ctx, []float64{100}, []float64{500}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err := g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-100), test.ShouldBeLessThan, 5)

	// Move to 250mm
	err = g.MoveToPosition(ctx, []float64{250}, []float64{500}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-250), test.ShouldBeLessThan, 5)

	// Move to 10mm (near min, but inside range to avoid triggering limit switch)
	err = g.MoveToPosition(ctx, []float64{10}, []float64{500}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-10), test.ShouldBeLessThan, 5)

	// Move to 290mm (near max, but inside range to avoid triggering limit switch)
	err = g.MoveToPosition(ctx, []float64{290}, []float64{500}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-290), test.ShouldBeLessThan, 5)
}

func TestIntegrationOutOfRange(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewPrinterXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("printer-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	_, err = g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)

	err = g.MoveToPosition(ctx, []float64{-1}, []float64{500}, nil)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "out of range")

	err = g.MoveToPosition(ctx, []float64{301}, []float64{500}, nil)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "out of range")
}

func TestIntegrationReconfigureWithoutRehoming(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewPrinterXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("printer-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	_, err = g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)

	sa := g.(*singleAxis)
	savedLimits := make([]float64, len(sa.positionLimits))
	copy(savedLimits, sa.positionLimits)
	savedRange := sa.positionRange

	// Reconfigure with only speed changed (same motor/board/pins)
	newRig := testrig.NewPrinterXAxis()
	newRig.GantryMmPerSec = 600
	newCfg := rigConfig("printer-x", newRig)
	err = g.Reconfigure(ctx, newRig.Dependencies(), newCfg)
	test.That(t, err, test.ShouldBeNil)

	test.That(t, sa.positionLimits, test.ShouldResemble, savedLimits)
	test.That(t, sa.positionRange, test.ShouldEqual, savedRange)

	pos, err := g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, len(pos), test.ShouldEqual, 1)
}

func TestIntegrationNotHomedErrors(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewPrinterXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("printer-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	// Skip Home — operations should fail
	err = g.MoveToPosition(ctx, []float64{100}, []float64{500}, nil)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "is homed")

	_, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "is homed")
}

func TestIntegrationHomeTwoSwitchCNC(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewCNCMillXAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("cnc-x", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	homed, err := g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homed, test.ShouldBeTrue)

	sa := g.(*singleAxis)
	test.That(t, sa.positionLimits[0], test.ShouldBeBetween, -1.0, 1.0)
	test.That(t, sa.positionLimits[1], test.ShouldBeBetween, 119.0, 121.0)

	pos, err := g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pos[0], test.ShouldBeBetween, 270, 330) // center ~300mm

	// Move to a position and verify
	err = g.MoveToPosition(ctx, []float64{100}, []float64{166}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-100), test.ShouldBeLessThan, 5)
}

func TestIntegrationHomeAndMoveSingleSwitch(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	rig := testrig.NewSingleSwitchAxis()

	g, err := newSingleAxis(ctx, rig.Dependencies(), rigConfig("single-sw", rig), logger)
	test.That(t, err, test.ShouldBeNil)

	defer g.Close(ctx)

	homed, err := g.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homed, test.ShouldBeTrue)

	// After homing, position should be at center
	pos, err := g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pos[0], test.ShouldBeBetween, 450, 550) // center ~500mm

	// Move to a specific position
	err = g.MoveToPosition(ctx, []float64{200}, []float64{200}, nil)
	test.That(t, err, test.ShouldBeNil)
	pos, err = g.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-200), test.ShouldBeLessThan, 5)
}

// Ensure integration tests don't break the Gantry interface contract.
var _ gantry.Gantry = (*singleAxis)(nil)
