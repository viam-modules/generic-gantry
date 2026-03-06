package multiaxis

import (
	"context"
	"math"
	"testing"

	"go.viam.com/test"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/gantry"
	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/testutils/inject"

	"github.com/viam-modules/generic-gantry/singleaxis"
	"github.com/viam-modules/generic-gantry/testrig"
)

// buildSingleAxis creates a real singleAxis gantry from a TestRig using the registered constructor.
func buildSingleAxis(
	ctx context.Context,
	t *testing.T,
	name string,
	rig *testrig.TestRig,
	logger logging.Logger,
) gantry.Gantry {
	t.Helper()

	saCfg := &singleaxis.Config{
		Motor:           rig.MotorName,
		LengthMm:        rig.LengthMm,
		MmPerRevolution: rig.MmPerRevolution,
		GantryMmPerSec:  rig.GantryMmPerSec,
	}
	if len(rig.LimitSwitchPins) > 0 {
		saCfg.Board = rig.BoardName
		saCfg.LimitSwitchPins = rig.LimitSwitchPins
		saCfg.LimitPinEnabled = &rig.LimitPinEnabled
	}

	cfg := resource.Config{
		Name:                name,
		ConvertedAttributes: saCfg,
	}

	deps := make(resource.Dependencies)
	deps[motor.Named(rig.MotorName)] = rig.Motor
	if rig.Board != nil {
		deps[board.Named(rig.BoardName)] = rig.Board
	}

	reg, ok := resource.LookupRegistration(gantry.API, singleaxis.Model)
	test.That(t, ok, test.ShouldBeTrue)

	g, err := reg.Constructor(ctx, deps, cfg, logger)
	test.That(t, err, test.ShouldBeNil)

	return g.(gantry.Gantry)
}

// wrapWithKinematics wraps a real gantry with an inject that provides a simple kinematics model,
// since singleAxis without a kinematics file returns nil from Kinematics().
func wrapWithKinematics(name string, real gantry.Gantry) *inject.Gantry {
	wrapped := inject.NewGantry(name)
	wrapped.PositionFunc = func(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
		return real.Position(ctx, extra)
	}
	wrapped.MoveToPositionFunc = func(ctx context.Context, pos, speed []float64, extra map[string]interface{}) error {
		return real.MoveToPosition(ctx, pos, speed, extra)
	}
	wrapped.LengthsFunc = func(ctx context.Context, extra map[string]interface{}) ([]float64, error) {
		return real.Lengths(ctx, extra)
	}
	wrapped.StopFunc = func(ctx context.Context, extra map[string]interface{}) error {
		return real.Stop(ctx, extra)
	}
	wrapped.HomeFunc = func(ctx context.Context, extra map[string]interface{}) (bool, error) {
		return real.Home(ctx, extra)
	}
	wrapped.CloseFunc = func(ctx context.Context) error {
		return real.Close(ctx)
	}
	wrapped.KinematicsFunc = func(ctx context.Context) (referenceframe.Model, error) {
		return referenceframe.NewSimpleModel(name), nil
	}
	return wrapped
}

func TestIntegrationMultiAxisHomingAndMove(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)

	rigX := testrig.NewPrinterXAxis()
	rigY := testrig.NewPrinterXAxis()
	rigY.Motor = testrig.NewSimulatedMotor("motor2", 4.0)
	rigY.MotorName = "motor2"
	rigY.Board = testrig.NewSimulatedBoard("board2", map[string]*testrig.LimitSwitchPin{
		"pin0": testrig.NewLimitSwitchPin(rigY.Motor, 0.0, false, true),
		"pin1": testrig.NewLimitSwitchPin(rigY.Motor, 7.5, true, true),
	})
	rigY.BoardName = "board2"

	realX := buildSingleAxis(ctx, t, "axis-x", rigX, logger)
	defer realX.Close(ctx)
	realY := buildSingleAxis(ctx, t, "axis-y", rigY, logger)
	defer realY.Close(ctx)

	homedX, err := realX.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homedX, test.ShouldBeTrue)

	homedY, err := realY.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, homedY, test.ShouldBeTrue)

	// Wrap with kinematics for multiaxis constructor
	axisX := wrapWithKinematics("axis-x", realX)
	axisY := wrapWithKinematics("axis-y", realY)

	deps := make(resource.Dependencies)
	deps[gantry.Named("axis-x")] = axisX
	deps[gantry.Named("axis-y")] = axisY

	maCfg := resource.Config{
		Name: "multi-xy",
		ConvertedAttributes: &Config{
			SubAxes: []string{"axis-x", "axis-y"},
		},
	}

	ma, err := newMultiAxis(ctx, deps, maCfg, logger)
	test.That(t, err, test.ShouldBeNil)
	defer ma.Close(ctx)

	// Both axes should be near center (~150mm each)
	pos, err := ma.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, len(pos), test.ShouldEqual, 2)
	test.That(t, pos[0], test.ShouldBeBetween, 120, 180)
	test.That(t, pos[1], test.ShouldBeBetween, 120, 180)

	// Move to specific positions
	err = ma.MoveToPosition(ctx, []float64{100, 200}, []float64{500, 500}, nil)
	test.That(t, err, test.ShouldBeNil)

	pos, err = ma.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, math.Abs(pos[0]-100), test.ShouldBeLessThan, 5)
	test.That(t, math.Abs(pos[1]-200), test.ShouldBeLessThan, 5)
}

func TestIntegrationMultiAxisStop(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)

	rigX := testrig.NewPrinterXAxis()
	rigY := testrig.NewPrinterXAxis()
	rigY.Motor = testrig.NewSimulatedMotor("motor2", 4.0)
	rigY.MotorName = "motor2"
	rigY.Board = testrig.NewSimulatedBoard("board2", map[string]*testrig.LimitSwitchPin{
		"pin0": testrig.NewLimitSwitchPin(rigY.Motor, 0.0, false, true),
		"pin1": testrig.NewLimitSwitchPin(rigY.Motor, 7.5, true, true),
	})
	rigY.BoardName = "board2"

	realX := buildSingleAxis(ctx, t, "axis-x", rigX, logger)
	defer realX.Close(ctx)
	realY := buildSingleAxis(ctx, t, "axis-y", rigY, logger)
	defer realY.Close(ctx)

	_, err := realX.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	_, err = realY.Home(ctx, nil)
	test.That(t, err, test.ShouldBeNil)

	axisX := wrapWithKinematics("axis-x", realX)
	axisY := wrapWithKinematics("axis-y", realY)

	deps := make(resource.Dependencies)
	deps[gantry.Named("axis-x")] = axisX
	deps[gantry.Named("axis-y")] = axisY

	maCfg := resource.Config{
		Name: "multi-xy",
		ConvertedAttributes: &Config{
			SubAxes: []string{"axis-x", "axis-y"},
		},
	}

	ma, err := newMultiAxis(ctx, deps, maCfg, logger)
	test.That(t, err, test.ShouldBeNil)
	defer ma.Close(ctx)

	err = ma.MoveToPosition(ctx, []float64{100, 200}, []float64{500, 500}, nil)
	test.That(t, err, test.ShouldBeNil)

	err = ma.Stop(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
}
