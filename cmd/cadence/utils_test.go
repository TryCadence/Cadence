package main

import (
	"testing"
)

func TestDetectFormatFromExtension(t *testing.T) {
	tests := []struct {
		filePath      string
		expected      string
		shouldError   bool
		errorContains string
	}{
		{
			filePath:    "report.json",
			expected:    "json",
			shouldError: false,
		},
		{
			filePath:    "report.txt",
			expected:    "text",
			shouldError: false,
		},
		{
			filePath:    "report.text",
			expected:    "text",
			shouldError: false,
		},
		{
			filePath:    "report",
			expected:    "text",
			shouldError: false,
		},
		{
			filePath:    "report.JSON",
			expected:    "json",
			shouldError: false,
		},
		{
			filePath:    "report.TXT",
			expected:    "text",
			shouldError: false,
		},
		{
			filePath:      "report.csv",
			expected:      "",
			shouldError:   true,
			errorContains: "unsupported file extension",
		},
		{
			filePath:      "report.pdf",
			expected:      "",
			shouldError:   true,
			errorContains: "unsupported file extension",
		},
		{
			filePath:    "/path/to/report.json",
			expected:    "json",
			shouldError: false,
		},
		{
			filePath:    "C:\\path\\to\\report.txt",
			expected:    "text",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			format, err := detectFormatFromExtension(tt.filePath)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				// Additional check for error message if needed
				if tt.errorContains != "" && err != nil {
					errMsg := err.Error()
					if errMsg == "" || len(tt.errorContains) > len(errMsg) {
						// Check if errorContains is in errMsg
						found := false
						for i := 0; i <= len(errMsg)-len(tt.errorContains); i++ {
							if errMsg[i:i+len(tt.errorContains)] == tt.errorContains {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("expected error containing %q, got %q", tt.errorContains, errMsg)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if format != tt.expected {
					t.Errorf("expected format %q, got %q", tt.expected, format)
				}
			}
		})
	}
}

func TestIsRemoteRepo(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{
			path:     "https://github.com/owner/repo",
			expected: true,
		},
		{
			path:     "http://github.com/owner/repo",
			expected: true,
		},
		{
			path:     "https://gitlab.com/owner/repo",
			expected: true,
		},
		{
			path:     "http://example.com/repo.git",
			expected: true,
		},
		{
			path:     "/home/user/repo",
			expected: false,
		},
		{
			path:     "C:\\Users\\repo",
			expected: false,
		},
		{
			path:     "https://",
			expected: false,
		},
		{
			path:     "http://",
			expected: false,
		},
		{
			path:     "",
			expected: false,
		},
		{
			path:     "ftp://example.com/repo",
			expected: false,
		},
		{
			path:     "./relative/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isRemoteRepo(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %q", tt.expected, result, tt.path)
			}
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		url            string
		expectedGitURL string
		expectedBranch string
	}{
		{
			url:            "https://github.com/owner/repo",
			expectedGitURL: "https://github.com/owner/repo.git",
			expectedBranch: "",
		},
		{
			url:            "https://github.com/owner/repo/blob/main/file.txt",
			expectedGitURL: "https://github.com/owner/repo.git",
			expectedBranch: "main",
		},
		{
			url:            "https://github.com/owner/repo/blob/develop/path/to/file.go",
			expectedGitURL: "https://github.com/owner/repo.git",
			expectedBranch: "develop",
		},
		{
			url:            "https://github.com/owner/repo/tree/feature/new-feature",
			expectedGitURL: "https://github.com/owner/repo.git",
			expectedBranch: "feature",
		},
		{
			url:            "https://gitlab.com/owner/repo",
			expectedGitURL: "https://gitlab.com/owner/repo",
			expectedBranch: "",
		},
		{
			url:            "http://example.com/repo",
			expectedGitURL: "http://example.com/repo",
			expectedBranch: "",
		},
		{
			url:            "https://github.com/owner",
			expectedGitURL: "https://github.com/owner",
			expectedBranch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			gitURL, branch := parseGitHubURL(tt.url)

			if gitURL != tt.expectedGitURL {
				t.Errorf("expected git URL %q, got %q", tt.expectedGitURL, gitURL)
			}
			if branch != tt.expectedBranch {
				t.Errorf("expected branch %q, got %q", tt.expectedBranch, branch)
			}
		})
	}
}
