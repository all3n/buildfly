package config

import (
	"testing"
)

func TestSourceInfo_GetURLs(t *testing.T) {
	tests := []struct {
		name     string
		source   SourceInfo
		expected []string
	}{
		{
			name: "multiple URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{
					"https://example.com/file1.tar.gz",
					"https://example.com/file2.tar.gz",
				},
			},
			expected: []string{
				"https://example.com/file1.tar.gz",
				"https://example.com/file2.tar.gz",
			},
		},
		{
			name: "single URL",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{"https://example.com/file.tar.gz"},
			},
			expected: []string{"https://example.com/file.tar.gz"},
		},
		{
			name: "empty URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{},
			},
			expected: []string{},
		},
		{
			name: "nil URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: nil,
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.source.GetURLs()
			if len(result) != len(tt.expected) {
				t.Errorf("GetURLs() returned %d URLs, expected %d", len(result), len(tt.expected))
				return
			}
			for i, url := range result {
				if url != tt.expected[i] {
					t.Errorf("GetURLs()[%d] = %s, expected %s", i, url, tt.expected[i])
				}
			}
		})
	}
}

func TestSourceInfo_HasURLs(t *testing.T) {
	tests := []struct {
		name     string
		source   SourceInfo
		expected bool
	}{
		{
			name: "has URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{"https://example.com/file.tar.gz"},
			},
			expected: true,
		},
		{
			name: "empty URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{},
			},
			expected: false,
		},
		{
			name: "nil URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.source.HasURLs()
			if result != tt.expected {
				t.Errorf("HasURLs() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSourceInfo_GetPrimaryURL(t *testing.T) {
	tests := []struct {
		name     string
		source   SourceInfo
		expected string
	}{
		{
			name: "has URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{
					"https://example.com/file1.tar.gz",
					"https://example.com/file2.tar.gz",
				},
			},
			expected: "https://example.com/file1.tar.gz",
		},
		{
			name: "empty URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: []string{},
			},
			expected: "",
		},
		{
			name: "nil URLs",
			source: SourceInfo{
				Type: "archive",
				URLS: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.source.GetPrimaryURL()
			if result != tt.expected {
				t.Errorf("GetPrimaryURL() = %s, expected %s", result, tt.expected)
			}
		})
	}
}
