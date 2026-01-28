package detector

import "fmt"

type Thresholds struct {
	// Size-based detection
	SuspiciousAdditions int64
	SuspiciousDeletions int64

	// Velocity-based detection
	MaxAdditionsPerMin float64
	MaxDeletionsPerMin float64

	// Timing-based detection
	MinTimeDeltaSeconds int64

	// File dispersion (batch changes)
	MaxFilesPerCommit int

	// Ratio-based detection (AI patterns)
	MaxAdditionRatio   float64 // Default: 0.95 (95% additions)
	MinDeletionRatio   float64 // Default: 0.95 (95% deletions)
	MinCommitSizeRatio int64   // Minimum commit size to apply ratio analysis

	// Precision/consistency detection
	EnablePrecisionAnalysis bool
}

func (t *Thresholds) Validate() error {
	if t.SuspiciousAdditions < 0 {
		return fmt.Errorf("SuspiciousAdditions cannot be negative")
	}

	if t.SuspiciousDeletions < 0 {
		return fmt.Errorf("SuspiciousDeletions cannot be negative")
	}

	if t.MaxAdditionsPerMin < 0 {
		return fmt.Errorf("MaxAdditionsPerMin cannot be negative")
	}

	if t.MaxDeletionsPerMin < 0 {
		return fmt.Errorf("MaxDeletionsPerMin cannot be negative")
	}

	if t.MinTimeDeltaSeconds < 0 {
		return fmt.Errorf("MinTimeDeltaSeconds cannot be negative")
	}

	if t.MaxFilesPerCommit < 0 {
		return fmt.Errorf("MaxFilesPerCommit cannot be negative")
	}

	if t.MaxAdditionRatio < 0 || t.MaxAdditionRatio > 1.0 {
		return fmt.Errorf("MaxAdditionRatio must be between 0.0 and 1.0")
	}

	if t.MinDeletionRatio < 0 || t.MinDeletionRatio > 1.0 {
		return fmt.Errorf("MinDeletionRatio must be between 0.0 and 1.0")
	}

	if t.SuspiciousAdditions == 0 &&
		t.SuspiciousDeletions == 0 &&
		t.MaxAdditionsPerMin == 0 &&
		t.MaxDeletionsPerMin == 0 &&
		t.MinTimeDeltaSeconds == 0 &&
		t.MaxFilesPerCommit == 0 &&
		t.MaxAdditionRatio == 0 &&
		t.MinDeletionRatio == 0 {
		return fmt.Errorf("at least one threshold must be configured")
	}

	return nil
}

func (t *Thresholds) IsZero() bool {
	return t.SuspiciousAdditions == 0 &&
		t.SuspiciousDeletions == 0 &&
		t.MaxAdditionsPerMin == 0 &&
		t.MaxDeletionsPerMin == 0 &&
		t.MinTimeDeltaSeconds == 0 &&
		t.MaxFilesPerCommit == 0 &&
		t.MaxAdditionRatio == 0 &&
		t.MinDeletionRatio == 0
}
