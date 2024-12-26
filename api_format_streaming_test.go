package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"
)

func createHost(config *Config) string {
	return config.Host + ":" + config.Port
}

func TestQuotesEndpoint(t *testing.T) {

	testConfig, srv := startTestServer("8001")
	defer func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			t.Errorf("Error shutting down server: %v", err)
		}
	}()

	time.Sleep(time.Second)

	testCases := []struct {
		Url          string
		ExpectedData string
	}{
		{"http://" + createHost(testConfig) + "/quotes/?page_size=1&format=json", "./test_data/expected_page_size_1.json"},
		{"http://" + createHost(testConfig) + "/quotes/?page_size=2&format=json", "./test_data/expected_page_size_2.json"},
		{"http://" + createHost(testConfig) + "/quotes/?page_size=3&format=json", "./test_data/expected_page_size_3.json"},
		{"http://" + createHost(testConfig) + "/quotes/?page_size=10&format=json", "./test_data/expected_page_size_10.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.Url, func(t *testing.T) {
			resp, err := http.Get(tc.Url)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

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
				t.Errorf("Response does not match expected result for %s\nGOT:\n%s\nExpected\n%s", tc.Url, actualJSON, expectedJSON)

			}
		})
	}
}

func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}

func startTestServer(port string) (*Config, *http.Server) {
	config := &Config{
		Filename:        "data/quotes.bytesz",
		Storage:         "bytesz",
		Port:            port,
		Host:            "127.0.0.1",
		DefaultPageSize: 10,
		MaxPageSize:     1000000,
		EnableLogging:   false,
	}

	quotes, err := LoadQuotes(config.Filename, config.Storage)
	if err != nil {
		panic(fmt.Sprintf("Error loading quotes: %v", err))
	}

	authorIndex := BuildAuthorIndex(quotes)
	tagIndex := BuildTagIndex(quotes)

	api := &API{
		Quotes:          quotes,
		Authors:         authorIndex,
		Tags:            tagIndex,
		DefaultPageSize: config.DefaultPageSize,
		MaxPageSize:     config.MaxPageSize,
		Runtime:         runtime.GOOS,
		EnableLogging:   config.EnableLogging,
	}
	mux := http.NewServeMux()
	api.SetupRoutes(mux)

	srv := &http.Server{
		Addr:    createHost(config),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("ListenAndServe(): %v", err))
		}
	}()

	return config, srv
}
