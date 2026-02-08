package metrics

import (
	"fmt"
	"time"
)

// MinTimeDelta is the minimum time delta used for velocity calculations.
// Any time delta shorter than this is clamped to this value to prevent
// inflated velocity metrics (e.g., 10 LOC in 1 second = 600 LOC/min).
const MinTimeDelta = 30 * time.Second

type VelocityMetrics struct {
	LOCPerMinute float64
	Clamped      bool // true if the time delta was clamped to MinTimeDelta
}

func CalculateVelocity(loc int64, timeDelta time.Duration) (*VelocityMetrics, error) {
	if timeDelta <= 0 {
		return nil, fmt.Errorf("invalid time delta: %v (must be positive)", timeDelta)
	}

	clamped := false
	if timeDelta < MinTimeDelta {
		timeDelta = MinTimeDelta
		clamped = true
	}

	velocity := float64(loc) / timeDelta.Minutes()
	return &VelocityMetrics{
		LOCPerMinute: velocity,
		Clamped:      clamped,
	}, nil
}

func CalculateVelocityPerMinute(loc int64, timeDelta time.Duration) (float64, error) {
	if timeDelta <= 0 {
		return 0, fmt.Errorf("invalid time delta: %v (must be positive)", timeDelta)
	}

	if timeDelta < MinTimeDelta {
		timeDelta = MinTimeDelta
	}

	return float64(loc) / timeDelta.Minutes(), nil
}
