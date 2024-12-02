package main

import (
	"testing"
)

func TestGetID(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectedID  int
		expectError bool
	}{
		// Happy path tests
		{
			name:        "Valid quote path",
			path:        "/quote/123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "Valid quote path with larger number",
			path:        "/quote/9876543",
			expectedID:  9876543,
			expectError: false,
		},

		// Happy path tests with trailing slash
		{
			name:        "Valid quote path with trailing slash",
			path:        "/quote/456/",
			expectedID:  456,
			expectError: false,
		},
		{
			name:        "Valid quote path with larger number and trailing slash",
			path:        "/quote/8765432/",
			expectedID:  8765432,
			expectError: false,
		},

		// Not happy path - expected failures
		{
			name:        "Invalid path format",
			path:        "/quote/abc",
			expectedID:  -1,
			expectError: true,
		},
		{
			name:        "Empty ID",
			path:        "/quote/",
			expectedID:  -1,
			expectError: true,
		},
		{
			name:        "Multiple slashes in path",
			path:        "/quote/123/456",
			expectedID:  456,
			expectError: false,
		},

		// Boundary tests
		{
			name:        "Minimum valid ID",
			path:        "/quote/0",
			expectedID:  0,
			expectError: false,
		},
		{
			name:        "Maximum int32 ID",
			path:        "/quote/2147483647",
			expectedID:  2147483647,
			expectError: false,
		},
		{
			name:        "Negative ID",
			path:        "/quote/-1",
			expectedID:  -1,
			expectError: true,
		},

		// Other cases
		{
			name:        "Path without quote",
			path:        "/other/path",
			expectedID:  -1,
			expectError: true,
		},
		{
			name:        "Path without quote with trailing slash",
			path:        "/other/path/",
			expectedID:  -1,
			expectError: true,
		},
		{
			name:        "Root path",
			path:        "/",
			expectedID:  -1,
			expectError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := getID(tt.path)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected an error, but got none", tt.name)
			}

			if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.name, err)
			}

			if id != tt.expectedID {
				t.Errorf("%s: Expected ID %d, but got %d", tt.name, tt.expectedID, id)
			}
		})
	}
}
