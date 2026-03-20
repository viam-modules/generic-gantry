// Package testrig provides simulated motor and board components for integration testing.
package testrig

import (
	"context"
	"math"
	"sync"
	"time"

	"go.viam.com/rdk/components/motor"
	"go.viam.com/rdk/resource"
)

// SimulatedMotor is a motor that tracks position based on time when RPM is active.
// Position is linearly interpolated: startPos + (rpm/60) * elapsed.
type SimulatedMotor struct {
	resource.Named
	resource.TriviallyCloseable
	resource.TriviallyReconfigurable

	mu        sync.Mutex
	position  float64   // frozen position (revolutions) when stopped
	rpm       float64   // current RPM (0 = stopped)
	startPos  float64   // position when RPM was last set
	startTime time.Time // when RPM was last set
	moving    bool      // true if SetRPM or GoTo active
	targetPos float64   // target for GoTo (only used when hasTarget)
	hasTarget bool      // true during GoTo, false during SetRPM
}

// NewSimulatedMotor creates a new SimulatedMotor at the given initial position.
func NewSimulatedMotor(name string, initialPosition float64) *SimulatedMotor {
	return &SimulatedMotor{
		Named:    motor.Named(name).AsNamed(),
		position: initialPosition,
	}
}

// currentPosition computes current position. Caller must hold m.mu.
func (m *SimulatedMotor) currentPosition() float64 {
	if !m.moving {
		return m.position
	}

	elapsed := time.Since(m.startTime).Seconds()

	pos := m.startPos + (m.rpm/60.0)*elapsed
	if m.hasTarget {
		if (m.rpm > 0 && pos >= m.targetPos) || (m.rpm < 0 && pos <= m.targetPos) {
			m.position = m.targetPos
			m.moving = false
			m.rpm = 0

			return m.targetPos
		}
	}

	return pos
}

func (m *SimulatedMotor) SetPower(ctx context.Context, powerPct float64, extra map[string]any) error {
	return nil
}

func (m *SimulatedMotor) SetRPM(ctx context.Context, rpm float64, extra map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cur := m.currentPosition()
	m.startPos = cur
	m.startTime = time.Now()
	m.rpm = rpm
	m.moving = rpm != 0
	m.hasTarget = false

	return nil
}

func (m *SimulatedMotor) GoFor(ctx context.Context, rpm, revolutions float64, extra map[string]any) error {
	m.mu.Lock()
	cur := m.currentPosition()
	m.mu.Unlock()

	return m.GoTo(ctx, rpm, cur+revolutions, extra)
}

func (m *SimulatedMotor) GoTo(ctx context.Context, rpm, targetPos float64, extra map[string]any) error {
	m.mu.Lock()
	cur := m.currentPosition()
	m.startPos = cur
	m.startTime = time.Now()
	m.targetPos = targetPos

	m.hasTarget = true
	if targetPos > cur {
		m.rpm = math.Abs(rpm)
	} else {
		m.rpm = -math.Abs(rpm)
	}

	m.moving = true
	m.mu.Unlock()

	// Block until target reached or context cancelled.
	for {
		err := ctx.Err()
		if err != nil {
			_ = m.Stop(ctx, nil)
			return err
		}

		m.mu.Lock()
		m.currentPosition() // updates state if target reached

		if !m.moving {
			// currentPosition() auto-stopped because target was reached
			m.mu.Unlock()
			return nil
		}
		m.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func (m *SimulatedMotor) ResetZeroPosition(ctx context.Context, offset float64, extra map[string]any) error {
	return nil
}

func (m *SimulatedMotor) Position(ctx context.Context, extra map[string]any) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.currentPosition(), nil
}

func (m *SimulatedMotor) Properties(ctx context.Context, extra map[string]any) (motor.Properties, error) {
	return motor.Properties{PositionReporting: true}, nil
}

func (m *SimulatedMotor) Stop(ctx context.Context, extra map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.position = m.currentPosition()
	m.moving = false
	m.rpm = 0

	return nil
}

func (m *SimulatedMotor) IsPowered(ctx context.Context, extra map[string]any) (bool, float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.moving, 0, nil
}

func (m *SimulatedMotor) IsMoving(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.moving, nil
}

func (m *SimulatedMotor) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, nil
}
