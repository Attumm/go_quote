package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestQuotesEndpoint_(t *testing.T) {
	testCases := []struct {
		Url          string
		ExpectedData string
	}{
		//{"http://127.0.0.1:8000/quotes/?page_size=0&format=json", "./test_data/expected_page_size_0.json"},
		{"http://127.0.0.1:8000/quotes/?page_size=1&format=json", "./test_data/expected_page_size_1.json"},
		{"http://127.0.0.1:8000/quotes/?page_size=2&format=json", "./test_data/expected_page_size_2.json"},
		{"http://127.0.0.1:8000/quotes/?page_size=3&format=json", "./test_data/expected_page_size_3.json"},
		{"http://127.0.0.1:8000/quotes/?page_size=10&format=json", "./test_data/expected_page_size_10.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.Url, func(t *testing.T) {
			resp, err := http.Get(tc.Url)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				if cerr := resp.Body.Close(); cerr != nil {
					t.Errorf("Failed to close response body: %v", cerr)
				}
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			expected, err := os.ReadFile(tc.ExpectedData)
			if err != nil {
				t.Fatalf("Failed to read expected result file: %v", err)
			}

			var actualJSON, expectedJSON interface{}
			if err := json.Unmarshal(body, &actualJSON); err != nil {
				t.Fatalf("Failed to parse actual JSON: %v", err)
			}
			if err := json.Unmarshal(expected, &expectedJSON); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}

			if !jsonEqual(actualJSON, expectedJSON) {
				t.Errorf("Response does not match expected result for %s", tc.Url)
			}
		})
	}
}

func jsonEqual_(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
