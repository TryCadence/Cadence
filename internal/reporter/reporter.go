package reporter

import (
	"fmt"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/reporter/formats"
)

type AnalysisFormatter interface {
	FormatAnalysis(report *analysis.AnalysisReport) (string, error)
}

func NewAnalysisFormatter(format string) (AnalysisFormatter, error) {
	switch format {
	case "text":
		return &formats.TextReporter{}, nil
	case "json":
		return &formats.JSONReporter{}, nil
	case "html":
		return &formats.HTMLReporter{}, nil
	case "yaml", "yml":
		return &formats.YAMLReporter{}, nil
	case "bson":
		return &formats.BSONReporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
}
