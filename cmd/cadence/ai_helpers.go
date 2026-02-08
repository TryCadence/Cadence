package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TryCadence/Cadence/internal/ai"
	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/config"
)

func performAIAnalysisUnified(report *analysis.AnalysisReport, aiConfig *config.AIConfig) error {
	aiAnalyzer, err := ai.NewAnalyzer(&ai.Config{
		Enabled:   aiConfig.Enabled,
		Provider:  aiConfig.Provider,
		APIKey:    aiConfig.APIKey,
		Model:     aiConfig.Model,
		MaxTokens: 500,
	})
	if err != nil {
		return fmt.Errorf("failed to create AI analyzer: %w", err)
	}

	if !aiAnalyzer.IsConfigured() {
		return fmt.Errorf("AI analyzer not properly configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for i, detection := range report.Detections {
		if !detection.Detected {
			continue
		}

		if len(detection.Examples) == 0 {
			continue
		}

		fmt.Fprintf(os.Stderr, "  Analyzing detection %d/%d: %s...\n", i+1, len(report.Detections), detection.Strategy)

		codeSnippet := detection.Examples[0]
		analysis, err := aiAnalyzer.AnalyzeSuspiciousCode(ctx, detection.Strategy, codeSnippet)
		if err != nil {
			fmt.Fprintf(os.Stderr, "    Warning: AI analysis failed for %s: %v\n", detection.Strategy, err)
			continue
		}

		if analysis != "" {
			detection.Description = detection.Description + " - AI: " + analysis
		}
	}

	return nil
}
