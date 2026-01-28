package metrics

import (
	"fmt"
	"time"
)

type VelocityMetrics struct {
	LOCPerMinute float64
}

// CalculateVelocity calculates lines of code per minute from a LOC count and time delta.
// Returns a VelocityMetrics struct with LOCPerMinute field populated.
// Returns an error if timeDelta is not positive.
func CalculateVelocity(loc int64, timeDelta time.Duration) (*VelocityMetrics, error) {
	velocity, err := CalculateVelocityPerMinute(loc, timeDelta)
	if err != nil {
		return nil, err
	}
	return &VelocityMetrics{
		LOCPerMinute: velocity,
	}, nil
}

// CalculateVelocityPerMinute calculates and returns the velocity as a float64 (LOC/minute).
// This is the primary calculation function; CalculateVelocity wraps this.
// Returns an error if timeDelta is not positive.
func CalculateVelocityPerMinute(loc int64, timeDelta time.Duration) (float64, error) {
	if timeDelta <= 0 {
		return 0, fmt.Errorf("invalid time delta: %v (must be positive)", timeDelta)
	}
	return float64(loc) / timeDelta.Minutes(), nil
}
