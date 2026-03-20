package testrig

import (
	"context"
	"testing"
	"time"

	"go.viam.com/test"
)

func TestSimulatedMotorPosition(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 0.0)

	// SetRPM at 6000 RPM = 100 rev/sec, sleep 50ms → ~5 revolutions
	err := m.SetRPM(ctx, 6000, nil)
	test.That(t, err, test.ShouldBeNil)
	time.Sleep(50 * time.Millisecond)

	pos, err := m.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pos, test.ShouldBeGreaterThan, 2.0)
	test.That(t, pos, test.ShouldBeLessThan, 10.0)

	moving, err := m.IsMoving(ctx)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, moving, test.ShouldBeTrue)
}

func TestSimulatedMotorGoTo(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 0.0)

	// GoTo blocks until target reached
	err := m.GoTo(ctx, 60000, 5.0, nil) // 1000 rev/sec, target 5.0 → ~5ms
	test.That(t, err, test.ShouldBeNil)

	pos, err := m.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pos, test.ShouldEqual, 5.0)

	moving, err := m.IsMoving(ctx)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, moving, test.ShouldBeFalse)
}

func TestSimulatedMotorGoToReverse(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 10.0)

	err := m.GoTo(ctx, 60000, 5.0, nil)
	test.That(t, err, test.ShouldBeNil)

	pos, err := m.Position(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pos, test.ShouldEqual, 5.0)
}

func TestSimulatedMotorStop(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 0.0)

	err := m.SetRPM(ctx, 6000, nil)
	test.That(t, err, test.ShouldBeNil)
	time.Sleep(20 * time.Millisecond)

	err = m.Stop(ctx, nil)
	test.That(t, err, test.ShouldBeNil)

	pos1, _ := m.Position(ctx, nil)

	time.Sleep(20 * time.Millisecond)

	pos2, _ := m.Position(ctx, nil)

	// Position should be frozen after stop
	test.That(t, pos1, test.ShouldEqual, pos2)
	test.That(t, pos1, test.ShouldBeGreaterThan, 0.0)
}

func TestSimulatedMotorProperties(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 0.0)

	props, err := m.Properties(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, props.PositionReporting, test.ShouldBeTrue)
}

func TestLimitSwitchPinMin(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", -1.0) // below threshold of 0
	pin := NewLimitSwitchPin(m, 0.0, false, true)

	val, err := pin.Get(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldBeTrue) // pos <= threshold, limitHigh=true → true
}

func TestLimitSwitchPinMax(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 10.0) // above threshold of 7.5
	pin := NewLimitSwitchPin(m, 7.5, true, true)

	val, err := pin.Get(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldBeTrue) // pos >= threshold, limitHigh=true → true
}

func TestLimitSwitchPinNotTriggered(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", 3.0) // between limits
	pinMin := NewLimitSwitchPin(m, 0.0, false, true)
	pinMax := NewLimitSwitchPin(m, 7.5, true, true)

	val, err := pinMin.Get(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldBeFalse) // pos > threshold, isMax=false → not triggered

	val, err = pinMax.Get(ctx, nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldBeFalse) // pos < threshold, isMax=true → not triggered
}

func TestLimitSwitchPinPolarity(t *testing.T) {
	ctx := context.Background()
	m := NewSimulatedMotor("m", -1.0)
	pinHigh := NewLimitSwitchPin(m, 0.0, false, true)
	pinLow := NewLimitSwitchPin(m, 0.0, false, false)

	valHigh, _ := pinHigh.Get(ctx, nil)
	valLow, _ := pinLow.Get(ctx, nil)

	test.That(t, valHigh, test.ShouldBeTrue)
	test.That(t, valLow, test.ShouldBeFalse) // inverted polarity
}
