package testrig

import (
	"context"
	"time"

	"github.com/pkg/errors"
	pb "go.viam.com/api/component/board/v1"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/resource"
)

// LimitSwitchPin is a GPIO pin whose state is derived from a SimulatedMotor's position.
type LimitSwitchPin struct {
	motor     *SimulatedMotor
	threshold float64 // motor position (revolutions) at which switch triggers
	isMax     bool    // true: triggers when pos >= threshold; false: when pos <= threshold
	limitHigh bool    // polarity: true = high when triggered
}

// NewLimitSwitchPin creates a position-aware limit switch pin.
func NewLimitSwitchPin(motor *SimulatedMotor, threshold float64, isMax, limitHigh bool) *LimitSwitchPin {
	return &LimitSwitchPin{
		motor:     motor,
		threshold: threshold,
		isMax:     isMax,
		limitHigh: limitHigh,
	}
}

func (p *LimitSwitchPin) Get(ctx context.Context, extra map[string]any) (bool, error) {
	pos, err := p.motor.Position(ctx, nil)
	if err != nil {
		return false, err
	}

	triggered := (p.isMax && pos >= p.threshold) || (!p.isMax && pos <= p.threshold)
	if p.limitHigh {
		return triggered, nil
	}

	return !triggered, nil
}

func (p *LimitSwitchPin) Set(ctx context.Context, high bool, extra map[string]any) error {
	return nil
}

func (p *LimitSwitchPin) PWM(ctx context.Context, extra map[string]any) (float64, error) {
	return 0, nil
}

func (p *LimitSwitchPin) SetPWM(ctx context.Context, dutyCyclePct float64, extra map[string]any) error {
	return nil
}

func (p *LimitSwitchPin) PWMFreq(ctx context.Context, extra map[string]any) (uint, error) {
	return 0, nil
}

func (p *LimitSwitchPin) SetPWMFreq(ctx context.Context, freqHz uint, extra map[string]any) error {
	return nil
}

// SimulatedBoard is a board with position-aware limit switch pins.
type SimulatedBoard struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable

	pins map[string]*LimitSwitchPin
}

// NewSimulatedBoard creates a board with the given limit switch pins.
func NewSimulatedBoard(name string, pins map[string]*LimitSwitchPin) *SimulatedBoard {
	return &SimulatedBoard{
		Named: board.Named(name).AsNamed(),
		pins:  pins,
	}
}

func (b *SimulatedBoard) GPIOPinByName(name string) (board.GPIOPin, error) {
	pin, ok := b.pins[name]
	if !ok {
		return nil, errors.Errorf("pin %q not found", name)
	}

	return pin, nil
}

func (b *SimulatedBoard) AnalogByName(name string) (board.Analog, error) {
	return nil, errors.New("not supported")
}

func (b *SimulatedBoard) DigitalInterruptByName(name string) (board.DigitalInterrupt, error) {
	return nil, errors.New("not supported")
}

func (b *SimulatedBoard) AnalogNames() []string { return nil }

func (b *SimulatedBoard) DigitalInterruptNames() []string { return nil }

func (b *SimulatedBoard) SetPowerMode(
	ctx context.Context, mode pb.PowerMode, duration *time.Duration, extra map[string]any,
) error {
	return errors.New("not supported")
}

func (b *SimulatedBoard) StreamTicks(
	ctx context.Context, interrupts []board.DigitalInterrupt, ch chan board.Tick, extra map[string]any,
) error {
	return errors.New("not supported")
}

func (b *SimulatedBoard) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, nil
}
