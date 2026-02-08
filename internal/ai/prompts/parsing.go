package prompts

import "strings"

// AnalysisResult holds the structured output from an AI code analysis.
type AnalysisResult struct {
	Assessment string   // "likely AI-generated", "possibly AI-generated", "unlikely AI-generated"
	Confidence float64  // 0.0-1.0
	Reasoning  string   // Brief explanation of key indicators
	Indicators []string // Specific patterns detected
}

// ParseAnalysisResult extracts an AnalysisResult from raw AI response text.
// It attempts to find and parse JSON in the response, falling back to text heuristics.
func ParseAnalysisResult(responseText string) (*AnalysisResult, error) {
	result := &AnalysisResult{}

	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		result.Assessment, result.Confidence = GetAssessmentFromText(responseText)
		result.Reasoning = responseText[:intMin(len(responseText), 200)]
		return result, nil
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]

	result.Assessment, result.Confidence = GetAssessmentFromText(jsonStr)

	confStart := strings.Index(jsonStr, `"confidence":`)
	if confStart != -1 {
		confStart += 13
		confEnd := strings.IndexAny(jsonStr[confStart:], ",}")
		if confEnd != -1 {
			confStr := strings.TrimSpace(jsonStr[confStart : confStart+confEnd])
			result.Confidence = ParseConfidence(confStr)
		}
	}

	reasonStart := strings.Index(jsonStr, `"reasoning":`)
	if reasonStart != -1 {
		reasonStart += 12
		reasonStart = strings.Index(jsonStr[reasonStart:], `"`) + reasonStart + 1
		reasonEnd := strings.Index(jsonStr[reasonStart:], `"`) + reasonStart
		if reasonEnd > reasonStart {
			result.Reasoning = jsonStr[reasonStart:reasonEnd]
		}
	}

	return result, nil
}

// GetAssessmentFromText determines the assessment and default confidence from free text.
func GetAssessmentFromText(text string) (assessment string, confidence float64) {
	switch {
	case strings.Contains(text, "likely"):
		return "likely AI-generated", 0.8
	case strings.Contains(text, "possibly"):
		return "possibly AI-generated", 0.5
	default:
		return "unlikely AI-generated", 0.2
	}
}

// ParseConfidence converts a confidence string to a float64 value.
func ParseConfidence(confStr string) float64 {
	switch confStr {
	case "1", "1.0":
		return 1.0
	case "0", "0.0":
		return 0.0
	default:
		if confStr != "" && confStr[0:1] == "0" {
			return 0.5
		}
		return 0.5
	}
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
