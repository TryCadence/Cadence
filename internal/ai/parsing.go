package ai

import "github.com/TryCadence/Cadence/internal/ai/prompts"

// Re-export parsing types and functions from the prompts package
// so that existing consumers of the ai package continue to work.

// AnalysisResult holds the structured output from an AI code analysis.
type AnalysisResult = prompts.AnalysisResult

// ParseAnalysisResult extracts an AnalysisResult from raw AI response text.
var ParseAnalysisResult = prompts.ParseAnalysisResult

// GetAssessmentFromText determines the assessment and default confidence from free text.
var GetAssessmentFromText = prompts.GetAssessmentFromText

// ParseConfidence converts a confidence string to a float64 value.
var ParseConfidence = prompts.ParseConfidence
