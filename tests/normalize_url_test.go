package main

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "remove scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove www subdomain",
			inputURL: "http://www.example.com",
			expected: "example.com",
		},
		{
			name:     "remove trailing slash",
			inputURL: "http://example.com/",
			expected: "example.com",
		},
		{
			name:     "normalize path",
			inputURL: "http://example.com/path/",
			expected: "example.com/path",
		},
		{
			name:     "keep query parameters",
			inputURL: "https://example.com/path?query=123",
			expected: "example.com/path?query=123",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeURL(tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if actual != tc.expected {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
