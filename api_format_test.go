package main

import (
	"net/http"
	"testing"
)

func TestGetOutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		acceptHeader   string
		expectedFormat string
	}{
		// Happy paths
		{
			name:           "Default to JSON",
			url:            "http://go-quote.com",
			acceptHeader:   "",
			expectedFormat: "json",
		},
		{
			name:           "Format from URL",
			url:            "http://go-quote.com?format=xml",
			acceptHeader:   "",
			expectedFormat: "xml",
		},
		{
			name:           "HTML from Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "text/html",
			expectedFormat: "html",
		},
		{
			name:           "Markdown from Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "text/markdown",
			expectedFormat: "markdown",
		},
		{
			name:           "YAML from Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "application/x-yaml",
			expectedFormat: "yaml",
		},

		// Boundary cases
		{
			name:           "URL format overrides Accept header",
			url:            "http://go-quote.com?format=xml",
			acceptHeader:   "text/html",
			expectedFormat: "xml",
		},
		{
			name:           "Case insensitive Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "TEXT/HTML",
			expectedFormat: "html",
		},
		{
			name:           "Multiple Accept values",
			url:            "http://go-quote.com",
			acceptHeader:   "application/json, text/html",
			expectedFormat: "html",
		},
		{
			name:           "Unrecognized format in URL",
			url:            "http://go-quote.com?format=unknown",
			acceptHeader:   "",
			expectedFormat: "json",
		},
		{
			name:           "Unrecognized Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "application/unknown",
			expectedFormat: "json",
		},
		{
			name:           "HTML Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "text/html",
			expectedFormat: "html",
		},
		{
			name:           "XML Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "application/xml",
			expectedFormat: "xml",
		},
		{
			name:           "Plain text Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "text/plain",
			expectedFormat: "text",
		},
		{
			name:           "Markdown Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "text/markdown",
			expectedFormat: "markdown",
		},
		{
			name:           "YAML Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "application/x-yaml",
			expectedFormat: "yaml",
		},
		{
			name:           "Atom format parameter",
			url:            "http://go-quote.com?format=atom",
			acceptHeader:   "",
			expectedFormat: "atom",
		},
		{
			name:           "CSV format parameter",
			url:            "http://go-quote.com?format=csv",
			acceptHeader:   "",
			expectedFormat: "csv",
		},
		{
			name:           "Embed format parameter",
			url:            "http://go-quote.com?format=embed",
			acceptHeader:   "",
			expectedFormat: "embed",
		},
		{
			name:           "Embed.js format parameter",
			url:            "http://go-quote.com?format=embed.js",
			acceptHeader:   "",
			expectedFormat: "embed.js",
		},
		{
			name:           "HTML format parameter",
			url:            "http://go-quote.com?format=html",
			acceptHeader:   "",
			expectedFormat: "html",
		},
		{
			name:           "JSON format parameter",
			url:            "http://go-quote.com?format=json",
			acceptHeader:   "",
			expectedFormat: "json",
		},
		{
			name:           "Markdown format parameter",
			url:            "http://go-quote.com?format=markdown",
			acceptHeader:   "",
			expectedFormat: "markdown",
		},
		{
			name:           "OEmbed format parameter",
			url:            "http://go-quote.com?format=oembed",
			acceptHeader:   "",
			expectedFormat: "oembed",
		},
		{
			name:           "OEmbed XML format parameter",
			url:            "http://go-quote.com?format=oembed.xml",
			acceptHeader:   "",
			expectedFormat: "oembed.xml",
		},
		{
			name:           "RSS format parameter",
			url:            "http://go-quote.com?format=rss",
			acceptHeader:   "",
			expectedFormat: "rss",
		},
		{
			name:           "SVG format parameter",
			url:            "http://go-quote.com?format=svg",
			acceptHeader:   "",
			expectedFormat: "svg",
		},
		{
			name:           "SVG download format parameter",
			url:            "http://go-quote.com?format=svg-download",
			acceptHeader:   "",
			expectedFormat: "svg-download",
		},
		{
			name:           "Text format parameter",
			url:            "http://go-quote.com?format=text",
			acceptHeader:   "",
			expectedFormat: "text",
		},
		{
			name:           "WAV format parameter",
			url:            "http://go-quote.com?format=wav",
			acceptHeader:   "",
			expectedFormat: "wav",
		},
		{
			name:           "XML format parameter",
			url:            "http://go-quote.com?format=xml",
			acceptHeader:   "",
			expectedFormat: "xml",
		},
		{
			name:           "YAML format parameter",
			url:            "http://go-quote.com?format=yaml",
			acceptHeader:   "",
			expectedFormat: "yaml",
		},
		{
			name:           "Unrecognized Accept header",
			url:            "http://go-quote.com",
			acceptHeader:   "application/unknown",
			expectedFormat: "json",
		},
		// Format precedence over accept header
		{
			name:           "Format parameter overrides Accept header - XML over HTML",
			url:            "http://go-quote.com?format=xml",
			acceptHeader:   "text/html",
			expectedFormat: "xml",
		},
		{
			name:           "Format parameter overrides Accept header - JSON over XML",
			url:            "http://go-quote.com?format=json",
			acceptHeader:   "application/xml",
			expectedFormat: "json",
		},
		{
			name:           "Format parameter overrides Accept header - Markdown over JSON",
			url:            "http://go-quote.com?format=markdown",
			acceptHeader:   "application/json",
			expectedFormat: "markdown",
		},
		{
			name:           "Format parameter overrides Accept header - CSV over YAML",
			url:            "http://go-quote.com?format=csv",
			acceptHeader:   "application/x-yaml",
			expectedFormat: "csv",
		},
		{
			name:           "Format parameter overrides Accept header - SVG over Plain Text",
			url:            "http://go-quote.com?format=svg",
			acceptHeader:   "text/plain",
			expectedFormat: "svg",
		},
		{
			name:           "Unrecognized format parameter uses valid Accept header",
			url:            "http://go-quote.com?format=unknown",
			acceptHeader:   "application/json",
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", tt.url, nil)
			r.Header.Set("Accept", tt.acceptHeader)

			got := getOutputFormat(r)
			if got != tt.expectedFormat {
				t.Errorf("getOutputFormat() = %v, want %v", got, tt.expectedFormat)
			}
		})
	}
}
