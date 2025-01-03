package main

import (
	//"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var OutputFormats = map[string]string{
	"atom":         "application/atom+xml",
	"csv":          "text/csv",
	"embed":        "text/html",
	"embed.js":     "application/javascript",
	"html":         "text/html",
	"json":         "application/json",
	"markdown":     "text/markdown",
	"oembed":       "application/json+oembed",
	"oembed.xml":   "text/xml+oembed",
	"rss":          "application/rss+xml",
	"svg":          "image/svg+xml",
	"svg-download": "image/svg+xml",
	"text":         "text/plain",
	"wav":          "audio/wav",
	"xml":          "application/xml",
	"yaml":         "application/yaml",
	//"mp3":          "audio/mpeg",
	//"ogg":          "audio/ogg",
	//"aiff":         "audio/aiff",
}

func getFormatFromURL(urlPath string) string {
	u, err := url.Parse(urlPath)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", urlPath, err)
		return ""
	}

	format := u.Query().Get("format")

	if _, ok := OutputFormats[format]; ok {
		return format
	}
	return ""
}

func getOutputFormat(r *http.Request) string {

	if format := getFormatFromURL(r.URL.String()); format != "" {
		return format
	}

	// Default to JSON
	format := "json"

	accept := strings.ToLower(r.Header.Get("Accept"))
	if accept != "" {
		switch {
		case strings.Contains(accept, "text/html"):
			format = "html"
		case strings.Contains(accept, "application/xml"):
			format = "xml"
		case strings.Contains(accept, "text/plain"):
			format = "text"
		case strings.Contains(accept, "text/markdown"):
			format = "markdown"
		case strings.Contains(accept, "application/x-yaml"):
			format = "yaml"
		}
	}
	return format
}

type XMLQuote struct {
	XMLName xml.Name `xml:"response"`
	ID      int      `xml:"id"`
	Text    string   `xml:"text"`
	Author  string   `xml:"author"`
	Tags    []string `xml:"tags>tag"`
}

func setPaginationHeaders(w http.ResponseWriter, pagination Pagination) {
	w.Header().Set("Current-Page", strconv.Itoa(pagination.Page))
	w.Header().Set("Page-Size", strconv.Itoa(pagination.PageSize))
	w.Header().Set("Total-Count", strconv.Itoa(pagination.Total))
	w.Header().Set("Total-Pages", strconv.Itoa(pagination.Pages))
	if pagination.Next != "" {
		w.Header().Set("Next-Page", pagination.Next)
	}
}

func (api *API) formatResponseQuotes(w http.ResponseWriter, response PaginatedQuotesResponse, format string) {
	setPaginationHeaders(w, response.Pagination)

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		fmt.Fprint(w, quotesToCSV(response.Quotes))
	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, quotesToHTML(response, ""))
	case "text":
		w.Header().Set("Content-Type", "text/plain")
		for _, quote := range response.Quotes {
			fmt.Fprintf(w, "Quote: %s\nAuthor: %s\nTags: %s\nID: %d\n\n",
				quote.Text, quote.Author, strings.Join(quote.Tags, ", "), quote.ID)
		}
	case "markdown":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		for _, quote := range response.Quotes {
			fmt.Fprint(w, quoteToMarkdown(quote))
		}
	case "yaml":
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		fmt.Fprint(w, quotesToYAML(response.Quotes))
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (api *API) formatResponseQuote(w http.ResponseWriter, quote ResponseQuote, responseInfo *ResponseInfo) {
	switch responseInfo.Format {
	case "json":
		serveJSONQuote(w, quote, api, responseInfo)
	case "xml":
		serveXMLQuote(w, quote, api, responseInfo)
	case "html":
		serveHTMLQuote(w, quote, api, responseInfo)
	case "text":
		serveTextQuote(w, quote, api, responseInfo)
	case "markdown":
		serveMarkdownQuote(w, quote, api, responseInfo)
	case "yaml":
		serveYAMLQuote(w, quote, api, responseInfo)
	case "csv":
		serveCSVQuote(w, quote, api, responseInfo)
	case "rss":
		serveRSSQuote(w, quote, api, responseInfo)
	case "atom":
		serveAtomQuote(w, quote, api, responseInfo)
	case "oembed", "oembed.json":
		serveOEmbedJSONQuote(w, quote, api, responseInfo)
	case "oembed.xml":
		serveOEmbedXMLQuote(w, quote, api, responseInfo)
	case "embed":
		serveEmbedQuote(w, quote, api, responseInfo)
	case "embed.js":
		serveEmbedJSQuote(w, quote, api, responseInfo)
	case "svg":
		serveSVGQuote(w, quote, api, responseInfo)
	case "svg-download":
		serveSVGQuoteDownload(w, quote, api, responseInfo)
	case "wav":
		serveWavQuote(w, quote, api, responseInfo)
	case "aiff":
		serveAiffQuote(w, quote, api, responseInfo)
	case "mp3":
		serveMP3Quote(w, quote, api, responseInfo)
	case "ogg":
		serveOggQuote(w, quote, api, responseInfo)
	default:
		serveJSONQuote(w, quote, api, responseInfo)
	}
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
