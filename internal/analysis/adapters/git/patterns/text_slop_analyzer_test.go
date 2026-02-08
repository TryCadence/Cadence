package patterns

import (
	"testing"
)

func TestTextSlopAnalyzer_AnalyzeContent(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectConfidence bool
		minExpectedScore int
	}{
		{
			name: "AI-generated marketing content",
			content: `Our innovative approach leverages cutting-edge technology to provide transformative value. 
In today's world, it is important to note that we are committed to delivering best-in-class service. 
Furthermore, our revolutionary platform offers unprecedented opportunities for growth. 
In conclusion, we believe this paradigm shift represents the future of the industry.
Additionally, utilizing our solution ensures optimal results and maximizes stakeholder satisfaction.`,
			expectConfidence: true,
			minExpectedScore: 20,
		},
		{
			name: "human-written casual content",
			content: `Hey, so I was thinking about this problem the other day. 
You know how sometimes things just don't work the way you'd expect? Well, that happened to me too.
Anyway, I figured out a workaround that's kinda hacky but it does the job.
Not sure if it's the best solution, but it works for my use case.
Let me know if you run into the same issue!`,
			expectConfidence: false,
			minExpectedScore: 0,
		},
		{
			name:             "minimal content",
			content:          "Hello world this is some minimal content that is just barely enough words to meet the minimum requirement for analysis. This is a test of the content analyzer system that needs to have more text. We are adding additional content here to make sure we definitely have enough words now.",
			expectConfidence: false,
			minExpectedScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewTextSlopAnalyzer()
			result, err := analyzer.AnalyzeContent(tt.content)

			if err != nil {
				t.Fatalf("AnalyzeContent() error = %v", err)
			}

			if result == nil {
				t.Fatal("AnalyzeContent() returned nil result")
			}

			score := result.GetConfidenceScore()
			if tt.expectConfidence && score < tt.minExpectedScore {
				t.Errorf("Expected confidence >= %d, got %d", tt.minExpectedScore, score)
			}

			t.Logf("Content: %q -> Confidence: %d%%, Patterns: %d",
				tt.name, score, len(result.Patterns))
		})
	}
}

func TestTextSlopAnalyzer_DetectOverusedPhrases(t *testing.T) {
	analyzer := NewTextSlopAnalyzer()
	content := "In today's world, it is important to note that furthermore, in conclusion, we must leverage our innovative platform to provide transformative solutions. Our revolutionary system provides unprecedented value and synergy across all stakeholder touchpoints and business objectives through innovative methodologies and advanced technical capabilities in today's dynamic business environment now."

	result, err := analyzer.AnalyzeContent(content)
	if err != nil {
		t.Fatalf("AnalyzeContent() error = %v", err)
	}
	if result == nil {
		t.Fatalf("AnalyzeContent() returned nil result")
	}
	if len(result.Patterns) == 0 {
		t.Errorf("Expected to detect overused phrases, found none")
	}
}

func TestTextSlopAnalyzer_DetectGenericLanguage(t *testing.T) {
	analyzer := NewTextSlopAnalyzer()
	content := "We provide value to our stakeholders and optimize outcomes across markets worldwide. We maximize efficiency and optimize workflows for improved performance outcomes daily basis. Utilize our solution to ensure results and success. Our service leverages best practices to facilitate growth and drive engagement across all verticals and markets worldwide today."

	result, err := analyzer.AnalyzeContent(content)
	if err != nil {
		t.Fatalf("AnalyzeContent() error = %v", err)
	}
	if result == nil {
		t.Fatalf("AnalyzeContent() returned nil result")
	}
	if len(result.Patterns) == 0 {
		t.Errorf("Expected to detect generic language, found none")
	}
}

func TestTextSlopAnalyzer_ConfidenceScoring(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		expected int
	}{
		{"zero rate", 0.0, 0},
		{"quarter rate", 0.25, 25},
		{"half rate", 0.5, 50},
		{"full rate", 1.0, 100},
		{"over 100", 1.5, 100}, // Should cap at 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TextSlopResult{SuspicionRate: tt.rate}
			if got := result.GetConfidenceScore(); got != tt.expected {
				t.Errorf("GetConfidenceScore() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestTextSlopResult_Summary(t *testing.T) {
	result := &TextSlopResult{
		Patterns: []Pattern{
			{Type: "test_pattern", Severity: 0.5, Description: "Test description"},
		},
		SuspicionRate: 0.75,
	}

	summary := result.GetSummary()
	if summary == "" {
		t.Error("GetSummary() returned empty string")
	}

	if len(result.Patterns) == 0 {
		t.Error("Summary should include detected patterns")
	}
}
