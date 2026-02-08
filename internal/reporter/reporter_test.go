package reporter

import (
	"testing"
)

func TestNewAnalysisFormatter(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "creates text formatter",
			format:      "text",
			expectError: false,
		},
		{
			name:        "creates json formatter",
			format:      "json",
			expectError: false,
		},
		{
			name:        "creates html formatter",
			format:      "html",
			expectError: false,
		},
		{
			name:        "creates yaml formatter",
			format:      "yaml",
			expectError: false,
		},
		{
			name:        "creates yml formatter",
			format:      "yml",
			expectError: false,
		},
		{
			name:        "creates bson formatter",
			format:      "bson",
			expectError: false,
		},
		{
			name:        "invalid format",
			format:      "xml",
			expectError: true,
		},
		{
			name:        "empty format",
			format:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAnalysisFormatter(tt.format)
			if tt.expectError {
				if err == nil {
					t.Errorf("NewAnalysisFormatter() expected error but got none")
				}
				if got != nil {
					t.Errorf("NewAnalysisFormatter() expected nil on error, got %v", got)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAnalysisFormatter() unexpected error = %v", err)
			}
			if got == nil {
				t.Error("NewAnalysisFormatter() returned nil")
			}
		})
	}
}
