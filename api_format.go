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
	"xml":          "application/xml",
	"html":         "text/html",
	"json":         "application/json",
	"text":         "text/plain",
	"markdown":     "text/markdown",
	"yaml":         "application/yaml",
	"csv":          "text/csv",
	"rss":          "application/rss+xml",
	"atom":         "application/atom+xml",
	"oembed":       "application/json+oembed",
	"oembed.xml":   "text/xml+oembed",
	"embed":        "text/html",
	"embed.js":     "application/javascript",
	"svg":          "image/svg+xml",
	"svg-download": "image/svg+xml",
	"wav":          "audio/wav",
	//"mp3":          "audio/mpeg",
	//"ogg":          "audio/ogg",
	//"aiff":         "audio/aiff",
	//"anker":        "application/json",
	//"json1":        "application/json",
	//"json2":        "application/json",
	//"json3":        "application/json",
	//"json4":        "application/json",
	//"json5":        "application/json",
	//"json6":        "application/json",
	//"json7":        "application/json",
	//"json8":        "application/json",
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

	accept := r.Header.Get("Accept")
	if accept != "" {
		if strings.Contains(accept, "text/html") {
			format = "html"
		} else if strings.Contains(accept, "application/xml") {
			format = "xml"
		} else if strings.Contains(accept, "text/plain") {
			format = "text"
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

func (api *API) formatResponseQuotes(w http.ResponseWriter, response PaginatedQuotes, format string) {
	w.Header().Set("Current-Page", strconv.Itoa(response.Pagination.Page))
	w.Header().Set("Page-Size", strconv.Itoa(response.Pagination.PageSize))
	w.Header().Set("Total-Count", strconv.Itoa(response.Pagination.Total))
	w.Header().Set("Total-Pages", strconv.Itoa(response.Pagination.Pages))
	if response.Pagination.Next != "" {
		w.Header().Set("Next-Page", response.Pagination.Next)
	}

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
