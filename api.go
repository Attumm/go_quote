package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type API struct {
	Quotes          Quotes
	Authors         IndexStructure
	Tags            IndexStructure
	DefaultPageSize int
	MaxPageSize     int
}

const PAGESIZE = "page_size"

func (api *API) SetupRoutes() {
	http.HandleFunc("/docs/", api.HandleFormatDocs)

	http.HandleFunc("/tags", api.ListTagsHandler)
	http.HandleFunc("/tags/", api.TagQuotesHandler)

	http.HandleFunc("/authors", api.ListAuthorsHandler)
	http.HandleFunc("/authors/", api.AuthorQuotesHandler)

	http.HandleFunc("/quotes/", api.ListQuotesHandler)
	http.HandleFunc("/random-quote", api.QuoteHandler)
	http.HandleFunc("/quote/", api.QuoteHandler)
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		PrintMemUsage()
		fmt.Fprintf(w, "Debug information printed to console")
	})
	http.HandleFunc("/", api.QuoteHandler)
}

func (api *API) TagQuotesHandler(w http.ResponseWriter, r *http.Request) {
	tagName := r.URL.Path[len("/tags/"):]

	quoteIDs, exists := api.Tags.nameToIDs[tagName]
	if !exists {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(quoteIDs), page, pageSize)

	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(quoteIDs) {
		endIndex = len(quoteIDs)
	}

	quotes := make([]ResponseQuote, 0, endIndex-startIndex)
	for _, id := range quoteIDs[startIndex:endIndex] {
		quotes = append(quotes, api.Quotes[id].CreateResponseQuote(id))
	}

	response := struct {
		Quotes     []ResponseQuote `json:"quotes"`
		Pagination Pagination      `json:"pagination"`
	}{
		Quotes:     quotes,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(api.Tags.names), page, pageSize)

	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(api.Tags.names) {
		endIndex = len(api.Tags.names)
	}

	type TagResponse struct {
		Name        string `json:"name"`
		TagID       string `json:"tag-id"`
		TotalQuotes int    `json:"total-quotes"`
	}

	tags := make([]TagResponse, 0, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		tags = append(tags, TagResponse{
			Name:        api.Tags.names[i],
			TagID:       api.Tags.names[i],
			TotalQuotes: len(api.Tags.nameToIDs[api.Tags.names[i]]),
		})
	}

	response := struct {
		Tags       []TagResponse `json:"tags"`
		Pagination Pagination    `json:"pagination"`
	}{
		Tags:       tags,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) ListAuthorsHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(api.Authors.names), page, pageSize)

	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(api.Authors.names) {
		endIndex = len(api.Authors.names)
	}

	type AuthorResponse struct {
		Name        string `json:"name"`
		AuthorID    string `json:"author-id"`
		TotalQuotes int    `json:"total-quotes"`
	}

	authors := make([]AuthorResponse, 0, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		encodedName := api.Authors.names[i]
		decodedName, _ := url.QueryUnescape(encodedName)
		authors = append(authors, AuthorResponse{
			Name:        decodedName,
			AuthorID:    encodedName,
			TotalQuotes: len(api.Authors.nameToIDs[encodedName]),
		})
	}

	response := struct {
		Authors    []AuthorResponse `json:"authors"`
		Pagination Pagination       `json:"pagination"`
	}{
		Authors:    authors,
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) AuthorQuotesHandler(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Path[len("/authors/"):]

	quoteIDs, exists := api.Authors.nameToIDs[authorID]
	if !exists {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(quoteIDs), page, pageSize)

	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(quoteIDs) {
		endIndex = len(quoteIDs)
	}

	quotes := make([]ResponseQuote, 0, endIndex-startIndex)
	for _, id := range quoteIDs[startIndex:endIndex] {
		quotes = append(quotes, api.Quotes[id].CreateResponseQuote(id))
	}

	authorName, _ := url.QueryUnescape(authorID)

	response := struct {
		Author      string          `json:"author"`
		AuthorID    string          `json:"author-id"`
		TotalQuotes int             `json:"total-quotes"`
		Quotes      []ResponseQuote `json:"quotes"`
		Pagination  Pagination      `json:"pagination"`
	}{
		Author:      authorName,
		AuthorID:    authorID,
		TotalQuotes: len(api.Authors.nameToIDs[authorID]),
		Quotes:      quotes,
		Pagination:  pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) ListQuotesHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(api.Quotes), page, pageSize)

	startIndex := (pagination.Page - 1) * pagination.PageSize
	endIndex := startIndex + pagination.PageSize
	if endIndex > len(api.Quotes) {
		endIndex = len(api.Quotes)
	}

	response := struct {
		Quotes     []ResponseQuote `json:"quotes"`
		Pagination Pagination      `json:"pagination"`
	}{
		Pagination: pagination,
	}

	for i := startIndex; i < endIndex; i++ {
		response.Quotes = append(response.Quotes, api.Quotes[i].CreateResponseQuote(i))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) QuoteHandler(w http.ResponseWriter, r *http.Request) {
	if len(api.Quotes) == 0 {
		http.Error(w, "No quotes available", http.StatusNotFound)
		return
	}

	quoteID := rand.Intn(len(api.Quotes))
	var err error
	if strings.Contains(r.URL.Path, "quote") {
		quoteID, err = getID(r.URL.Path)
		if err != nil || quoteID < 0 || quoteID >= len(api.Quotes) {
			http.Error(w, "Quote not found", http.StatusNotFound)
			return
		}
	}
	quote := api.Quotes[quoteID].CreateResponseQuote(quoteID)

	responseInfo := getResponseInfo(r, quoteID)
	//fmt.Println(responseInfo)
	//fmt.Println(quote)
	api.formatResponse(w, quote, responseInfo)
}

func (api *API) HandleFormatDocs(w http.ResponseWriter, r *http.Request) {
	quoteID := 1
	quote := api.Quotes[quoteID].CreateResponseQuote(quoteID)

	type FormatExample struct {
		Name        string
		Format      string
		ContentType string
		Example     string
	}

	examples := make([]FormatExample, 0, len(OutputFormats))

	responseInfo := getResponseInfo(r, quoteID)
	for format, contentType := range OutputFormats {
		responseInfo.Format = format

		buf := &bytes.Buffer{}
		respWriter := &responseWriter{
			header: make(http.Header),
			buffer: buf,
		}

		api.formatResponse(respWriter, quote, responseInfo)

		examples = append(examples, FormatExample{
			Name:        format,
			Format:      format,
			ContentType: contentType,
			Example:     buf.String(),
		})
	}

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Format < examples[j].Format
	})

	const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Output Formats Documentation</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
            color: #333;
        }
        .format-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 2rem;
            margin: 2rem 0;
        }
        .format-card {
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 1.5rem;
            background: #f8f9fa;
        }
        .format-card h3 {
            margin-top: 0;
            color: #2563eb;
        }
        .try-links {
            margin: 1rem 0;
        }
        .try-links a {
            display: inline-block;
            padding: 0.5rem 1rem;
            background: #2563eb;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin-right: 0.5rem;
            margin-bottom: 0.5rem;
        }
        .try-links a:hover {
            background: #1d4ed8;
        }
        code {
            background: #e5e7eb;
            padding: 0.2em 0.4em;
            border-radius: 3px;
            font-size: 0.9em;
        }
        pre {
            background: #1f2937;
            color: #e5e7eb;
            padding: 1rem;
            border-radius: 8px;
            overflow-x: auto;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        .method {
            margin: 2rem 0;
            padding: 1.5rem;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
        }
        .example-preview {
            margin-top: 1rem;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            overflow: hidden;
        }
        .example-preview iframe {
            width: 100%;
            height: 200px;
            border: none;
            background: white;
        }
        .preview-tabs {
            display: flex;
            border-bottom: 1px solid #e5e7eb;
        }
        .preview-tab {
            padding: 0.5rem 1rem;
            cursor: pointer;
            background: none;
            border: none;
            border-bottom: 2px solid transparent;
        }
        .preview-tab.active {
            border-bottom-color: #2563eb;
            color: #2563eb;
        }
        .preview-content {
            padding: 1rem;
        }
    </style>
</head>
<body>
    <h1>API Output Formats</h1>
    <p>The API supports multiple output formats. You can request different formats in two ways:</p>

    <div class="method">
        <h2>Method 1: Using Format Parameter</h2>
        <p>Add <code>?format=TYPE</code> to your URL:</p>
        <pre>GET /quote/1?format=json</pre>
    </div>

    <div class="method">
        <h2>Method 2: Using Accept Header</h2>
        <p>Set the appropriate Accept header in your request:</p>
        <pre>curl -H "Accept: application/json" /quote/1</pre>
    </div>

    <h2>Available Formats</h2>
    <div class="format-grid">
        {{range .Examples}}
        <div class="format-card">
            <h3>{{.Name}}</h3>
            <p>Format: <code>{{.Format}}</code></p>
            <p>Content-Type: <code>{{.ContentType}}</code></p>
            
            <div class="try-links">
                <h4>Try it:</h4>
                <a href="/quote/1?format={{.Format}}" target="_blank">View {{.Format}} format</a>
            </div>

            <h4>Using format parameter:</h4>
            <pre>GET /quote/1?format={{.Format}}</pre>

            <h4>Using Accept header:</h4>
            <pre>curl -H "Accept: {{.ContentType}}" /quote/1</pre>

            <div class="example-preview">
                <div class="preview-tabs">
                    <button class="preview-tab active" onclick="showTab(this, 'preview-{{.Format}}')">Preview</button>
                    <button class="preview-tab" onclick="showTab(this, 'source-{{.Format}}')">Source</button>
                </div>
                <div id="preview-{{.Format}}" class="preview-content">
                    <iframe src="/quote/1?format={{.Format}}"></iframe>
                </div>
                <div id="source-{{.Format}}" class="preview-content" style="display: none;">
                    <pre>{{.Example}}</pre>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    <script>
    function showTab(btn, contentId) {

        const tabs = btn.parentElement.getElementsByClassName('preview-tab');
        for (let tab of tabs) {
            tab.classList.remove('active');
        }
        btn.classList.add('active');

        const container = btn.closest('.example-preview');
        const contents = container.getElementsByClassName('preview-content');
        for (let content of contents) {
            content.style.display = 'none';
        }
        document.getElementById(contentId).style.display = 'block';
    }
    </script>
</body>
</html>`

	data := struct {
		Examples []FormatExample
		BaseURL  string
		QuoteID  int
	}{
		Examples: examples,
		BaseURL:  responseInfo.BaseURL,
		QuoteID:  quoteID,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("formats").Parse(htmlTemplate))
	tmpl.Execute(w, data)
}

type responseWriter struct {
	header http.Header
	buffer *bytes.Buffer
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	// No-op
}
