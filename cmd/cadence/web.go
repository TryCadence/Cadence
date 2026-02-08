package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/detectors"
	"github.com/TryCadence/Cadence/internal/analysis/sources"
	"github.com/TryCadence/Cadence/internal/config"
	"github.com/TryCadence/Cadence/internal/reporter"
)

var (
	webURL     string
	verbose    bool
	outputFile string
	jsonFormat bool
)

var webCmd = &cobra.Command{
	Use:   "web <url>",
	Short: "Analyze a website for AI-generated content",
	Long: `Analyze website content to detect AI-generated text ("slop").

Examines content for common AI patterns including:
- Overused phrases and generic language
- Excessive structure and formatting
- Suspiciously perfect grammar
- Boilerplate text patterns
- Lack of nuance and specific details

Generates detailed reports showing exactly what content was flagged and why.
Each detected pattern includes:
- Severity score
- Detailed reasoning
- Specific examples found in the content
- Context showing where patterns appear

Supports both human-readable text output and machine-readable JSON format.

Examples:
  # Basic analysis with text report
  cadence web https://example.com

  # Generate JSON report and save to file
  cadence web https://example.com --json --output report.json

  # Verbose output with content quality metrics
  cadence web https://example.com --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runWebAnalyze,
}

func init() {
	webCmd.Flags().StringVarP(&webURL, "url", "u", "", "website URL to analyze")
	webCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed analysis information")
	webCmd.Flags().StringVarP(&outputFile, "output", "o", "", "write report to file (saved in reports/ directory)")
	webCmd.Flags().BoolVarP(&jsonFormat, "json", "j", false, "output in JSON format")
}

func runWebAnalyze(cmd *cobra.Command, args []string) error {
	url := args[0]

	fmt.Fprintf(os.Stderr, "Analyzing website content from %s...\n", url)

	source := sources.NewWebsiteSource(url)
	webDetector := detectors.NewWebDetector()
	runner := analysis.NewDefaultDetectionRunner()

	report, err := runner.Run(context.Background(), source, webDetector)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Analysis complete: %d detections found\n", report.DetectionCount)
	}

	cfgPath := configFile
	if cfgPath == "" {
		if _, err := os.Stat("cadence.yml"); err == nil {
			cfgPath = "cadence.yml"
		}
	}

	cfg, err := config.Load(cfgPath)
	if err == nil && cfg.AI.Enabled && report.DetectionCount > 0 {
		fmt.Fprintf(os.Stderr, "Performing AI analysis...\n")
		if err := performAIAnalysisUnified(report, &cfg.AI); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: AI analysis failed: %v\n", err)
		}
	}

	outputFormat := "text"
	if jsonFormat {
		outputFormat = "json"
	}

	formatter, err := reporter.NewAnalysisFormatter(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	reportStr, err := formatter.FormatAnalysis(report)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	if outputFile != "" {
		reportsDir := "reports"
		if err := os.MkdirAll(reportsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create reports directory: %w", err)
		}

		fullPath := filepath.Join(reportsDir, outputFile)
		if err := os.WriteFile(fullPath, []byte(reportStr), 0o600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", fullPath)
	} else {
		fmt.Println(reportStr)
	}

	return nil
}
