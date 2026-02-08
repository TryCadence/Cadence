package analysis

import (
	"context"
)

type SourceData struct {
	ID         string                 // Unique identifier for the source data (e.g., commit hash, URL, etc.)
	Type       string                 // Type of the source (e.g., "git", "web", "npm", "docker" etc.)
	RawContent interface{}            // The raw data fetched from the source (e.g., commit data, webhook payload, etc.)
	Metadata   map[string]interface{} // Additional metadata about the source (e.g., author, timestamp, etc.)
}

type AnalysisSource interface {
	Type() string
	Validate(ctx context.Context) error
	Fetch(ctx context.Context) (*SourceData, error)
}

type Detector interface {
	Detect(ctx context.Context, data *SourceData) ([]Detection, error)
}

type DetectionRunner interface {
	Run(ctx context.Context, source AnalysisSource, detectors ...Detector) (*AnalysisReport, error)
}
