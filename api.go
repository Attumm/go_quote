package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
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
	Runtime         string
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

func calculateSafeIndices(total int, pagination Pagination) (startIndex, endIndex, capacity int) {
	startIndex = (pagination.Page - 1) * pagination.PageSize
	if startIndex >= total {
		return 0, 0, 0
	}

	endIndex = startIndex + pagination.PageSize
	if endIndex > total {
		endIndex = total
	}

	capacity = endIndex - startIndex
	return startIndex, endIndex, capacity
}

type PaginatedQuotes struct {
	Quotes     []ResponseQuote `json:"quotes"`
	Pagination Pagination      `json:"pagination"`
}

type AuthorResponse struct {
	Name        string `json:"name"`
	AuthorID    string `json:"author-id"`
	TotalQuotes int    `json:"total-quotes"`
}

type PaginatedAuthors struct {
	Authors    []AuthorResponse `json:"authors"`
	Pagination Pagination       `json:"pagination"`
}

type TagResponse struct {
	Name        string `json:"name"`
	TagID       string `json:"tag-id"`
	TotalQuotes int    `json:"total-quotes"`
}

type PaginatedTags struct {
	Tags       []TagResponse `json:"tags"`
	Pagination Pagination    `json:"pagination"`
}

func (api *API) TagQuotesHandler(w http.ResponseWriter, r *http.Request) {
	tagName := r.URL.Path[len("/tags/"):]

	quoteIDs, exists := api.Tags.NameToQuotes[tagName]
	if !exists {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(quoteIDs), page, pageSize)
	startIndex, endIndex, capacity := calculateSafeIndices(len(quoteIDs), pagination)

	quotes := make([]ResponseQuote, 0, capacity)
	for _, id := range quoteIDs[startIndex:endIndex] {
		quotes = append(quotes, api.Quotes[id].CreateResponseQuote(id))
	}

	response := PaginatedQuotes{
		Quotes:     quotes,
		Pagination: pagination,
	}

	api.formatResponseQuotes(w, response, getOutputFormat(r))
}

func (api *API) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestDataList(r, api, TagsTypeRequest)

	tags := make([]TagResponse, 0, requestData.Total)

	for i := requestData.StartIndex; i < requestData.EndIndex; i++ {
		tags = append(tags, TagResponse{
			Name:        api.Tags.Names[i],
			TagID:       api.Tags.Names[i],
			TotalQuotes: len(api.Tags.NameToQuotes[api.Tags.Names[i]]),
		})
	}

	response := struct {
		Tags       []TagResponse `json:"tags"`
		Pagination Pagination    `json:"pagination"`
	}{
		Tags:       tags,
		Pagination: requestData.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) ListAuthorsHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestDataList(r, api, AuthorsTypeRequest)

	authors := make([]AuthorResponse, 0, requestData.Total)
	for i := requestData.StartIndex; i < requestData.EndIndex; i++ {
		encodedName := api.Authors.Names[i]
		decodedName, _ := url.QueryUnescape(encodedName)
		authors = append(authors, AuthorResponse{
			Name:        decodedName,
			AuthorID:    encodedName,
			TotalQuotes: len(api.Authors.NameToQuotes[encodedName]),
		})
	}

	response := struct {
		Authors    []AuthorResponse `json:"authors"`
		Pagination Pagination       `json:"pagination"`
	}{
		Authors:    authors,
		Pagination: requestData.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *API) AuthorQuotesHandler(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Path[len("/authors/"):]

	quoteIDs, exists := api.Authors.NameToQuotes[authorID]
	if !exists {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get(PAGESIZE))

	pagination := api.paginate(len(quoteIDs), page, pageSize)
	startIndex, endIndex, capacity := calculateSafeIndices(len(quoteIDs), pagination)

	quotes := make([]ResponseQuote, 0, capacity)
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
		TotalQuotes: len(api.Authors.NameToQuotes[authorID]),
		Quotes:      quotes,
		Pagination:  pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type RequestDataList struct {
	Gzip            bool
	Format          string
	Page            int
	PageSize        int
	Pagination      Pagination
	StartIndex      int
	EndIndex        int
	Total           int
	RequestCategory Category
}

func createRequestDataList(r *http.Request, api *API, category Category) *RequestDataList {
	urlParameters := r.URL.Query()
	page, _ := strconv.Atoi(urlParameters.Get("page"))
	pageSize, _ := strconv.Atoi(urlParameters.Get(PAGESIZE))

	var dataLen int
	switch category {
	case QuotesTypeRequest:
		dataLen = len(api.Quotes)
	case AuthorsTypeRequest:
		dataLen = api.Authors.Len()
	case TagsTypeRequest:
		dataLen = api.Tags.Len()
	default:
		log.Fatal("Invalid data type provided")
	}

	pagination := api.paginate(dataLen, page, pageSize)
	startIndex, endIndex, capacity := calculateSafeIndices(dataLen, pagination)

	gzip := urlParameters.Get("gzip") == "true" || strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip")
	return &RequestDataList{
		Gzip:       gzip,
		Format:     getOutputFormat(r),
		Page:       page,
		PageSize:   pageSize,
		Pagination: pagination,
		StartIndex: startIndex,
		EndIndex:   endIndex,
		Total:      capacity,
	}
}

type RequestData struct {
	Gzip   bool
	Format string
}

func createRequestData(r *http.Request, api *API) *RequestData {
	urlParameters := r.URL.Query()
	gzip := urlParameters.Get("gzip") == "true" || strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip")
	return &RequestData{
		Gzip:   gzip,
		Format: getOutputFormat(r),
	}
}

type ResponseInfo struct {
	Gzip     bool
	QuoteID  int
	BaseURL  string
	QuoteURL string
	Format   string
}

func getResponseInfo(r *http.Request, quoteID int, requestdata *RequestData) *ResponseInfo {
	baseURL := fmt.Sprintf("%s://%s", scheme(r), r.Host)
	return &ResponseInfo{
		Gzip:     requestdata.Gzip,
		Format:   requestdata.Format,
		QuoteID:  quoteID,
		BaseURL:  baseURL,
		QuoteURL: fmt.Sprintf("%s/quote/%d", baseURL, quoteID),
	}
}

func (api *API) ListQuotesHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestDataList(r, api, QuotesTypeRequest)
	api.formatStreamingResponse(w, requestData)
}

func (api *API) QuoteHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestData(r, api)
	if len(api.Quotes) == 0 {
		http.Error(w, "No quotes available", http.StatusNotFound)
		return
	}

	quoteID := rand.Intn(len(api.Quotes))
	var err error
	if strings.Contains(r.URL.Path, "quote/") {
		quoteID, err = getID(r.URL.Path)
		if err != nil || quoteID < 0 || quoteID >= len(api.Quotes) {
			http.Error(w, "Quote not found", http.StatusNotFound)
			return
		}
	}
	quote := api.Quotes[quoteID].CreateResponseQuote(quoteID)

	responseInfo := getResponseInfo(r, quoteID, requestData)
	api.formatResponseQuote(w, quote, responseInfo)
}

func (api *API) HandleFormatDocs(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestData(r, api)
	quoteID := 1
	quote := api.Quotes[quoteID].CreateResponseQuote(quoteID)

	type FormatExample struct {
		Name        string
		Format      string
		ContentType string
		Example     string
		IsAudio     bool
	}

	examples := make([]FormatExample, 0, len(OutputFormats))

	responseInfo := getResponseInfo(r, quoteID, requestData)
	for format, contentType := range OutputFormats {
		responseInfo.Format = format

		buf := &bytes.Buffer{}
		respWriter := &responseWriter{
			header: make(http.Header),
			buffer: buf,
		}

		api.formatResponseQuote(respWriter, quote, responseInfo)

		isAudio := format == "ogg" || format == "mp3" || format == "aiff" || format == "wav"

		examples = append(examples, FormatExample{
			Name:        format,
			Format:      format,
			ContentType: contentType,
			Example:     buf.String(),
			IsAudio:     isAudio,
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
        audio {
            width: 100%;
            margin-top: 1rem;
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
                    {{if .IsAudio}}
                        <audio controls>
                            <source src="/quote/1?format={{.Format}}" type="{{.ContentType}}">
                            Your browser does not support the audio element.
                        </audio>
                    {{else}}
                        <iframe src="/quote/1?format={{.Format}}"></iframe>
                    {{end}}
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
