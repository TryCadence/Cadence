package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/TryCadence/Cadence/internal/ai"
	"github.com/TryCadence/Cadence/internal/config"
	"github.com/TryCadence/Cadence/internal/detector/patterns"
	"github.com/TryCadence/Cadence/internal/web"
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
	rootCmd.AddCommand(webCmd)
}

func runWebAnalyze(cmd *cobra.Command, args []string) error {
	url := args[0]

	fmt.Fprintf(os.Stderr, "Fetching website content from %s...\n", url)

	// Fetch website
	fetcher := web.NewFetcher(10 * time.Second)
	pageContent, err := fetcher.Fetch(url)
	if err != nil {
		return fmt.Errorf("failed to fetch website: %w", err)
	}

	// Check content quality
	quality := pageContent.GetContentQuality()
	if verbose {
		fmt.Fprintf(os.Stderr, "Content extracted: %d words, %d headings\n", pageContent.WordCount, len(pageContent.Headings))
		fmt.Fprintf(os.Stderr, "Content quality score: %.2f\n", quality)
	}

	if pageContent.WordCount < 50 {
		fmt.Fprintf(os.Stderr, "Warning: Content may be too short for reliable analysis (%d words)\n", pageContent.WordCount)
	}

	if quality < 0.5 {
		fmt.Fprintf(os.Stderr, "Warning: Low content quality detected (score: %.2f) - results may be less reliable\n", quality)
	}

	fmt.Fprintf(os.Stderr, "Analyzing content for AI patterns...\n")

	// Analyze for text slop
	analyzer := patterns.NewTextSlopAnalyzer()
	result, err := analyzer.AnalyzeContent(pageContent.GetMainContent())
	if err != nil {
		if pageContent.WordCount < 50 {
			return fmt.Errorf("content too short for analysis: %w (try analyzing a page with more text content)", err)
		}
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Perform AI analysis if enabled
	var aiAnalysis string
	cfgPath := configFile
	if cfgPath == "" {
		if _, err := os.Stat("cadence.yml"); err == nil {
			cfgPath = "cadence.yml"
		}
	}

	cfg, err := config.Load(cfgPath)
	if err == nil && cfg.AI.Enabled {
		fmt.Fprintf(os.Stderr, "Performing AI analysis...\n")
		aiAnalysis, err = performWebAIAnalysis(pageContent, result, &cfg.AI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: AI analysis failed: %v\n", err)
		}
	}

	// Generate report using new reporting system
	reportData := &web.WebReportData{
		Content:    pageContent,
		Analysis:   result,
		AIAnalysis: aiAnalysis,
		AnalyzedAt: time.Now(),
	}

	var reporter web.WebReporter
	if jsonFormat {
		reporter = &web.JSONWebReporter{}
	} else {
		reporter = &web.TextWebReporter{}
	}

	output, err := reporter.Generate(reportData)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Write output
	if outputFile != "" {
		// Create reports directory if it doesn't exist
		reportsDir := "reports"
		if err := os.MkdirAll(reportsDir, 0o750); err != nil {
			return fmt.Errorf("failed to create reports directory: %w", err)
		}

		// Construct full path
		fullPath := fmt.Sprintf("%s/%s", reportsDir, outputFile)

		if err := os.WriteFile(fullPath, []byte(output), 0o600); err != nil {
			return fmt.Errorf("failed to write report to file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", fullPath)
	} else {
		fmt.Print(output)
	}

	return nil
}

func performWebAIAnalysis(content *web.PageContent, result *patterns.TextSlopResult, aiCfg *config.AIConfig) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	analyzer, err := ai.NewOpenAIAnalyzer(aiCfg.APIKey, aiCfg.Model)
	if err != nil {
		return "", err
	}

	// Create analysis prompt for web content
	prompt := fmt.Sprintf(`Analyze the following website content for signs of AI generation. Consider the overall writing style, patterns, and content quality.

Website URL: %s
Title: %s
Content Summary: %d words

Content excerpt:
%s

Detected patterns: %d (confidence: %d%%)
- %v

Provide a brief assessment (2-3 sentences) of whether this content appears to be AI-generated.`,
		content.URL,
		content.Title,
		len(content.GetMainContent()),
		truncateText(content.GetMainContent(), 1000),
		len(result.Patterns),
		result.GetConfidenceScore(),
		result.Patterns,
	)

	// Use the OpenAI analyzer (reusing existing functionality)
	analysis, err := analyzer.AnalyzeWithSystemPrompt(ctx,
		"You are an expert at detecting AI-generated text and content. Analyze the provided website content and assess the likelihood it was generated by AI.",
		prompt,
	)
	if err != nil {
		return "", err
	}

	return analysis, nil
}

func truncateText(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "..."
}
