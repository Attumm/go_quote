package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
)

const PAGESIZE = "page_size"

type API struct {
	Quotes          Quotes
	Authors         IndexStructure
	Tags            IndexStructure
	DefaultPageSize int
	MaxPageSize     int
	Runtime         string
	EnableLogging   bool
	PermissiveCORS  bool
	Swagger         bool
}

func (api *API) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		URL := r.URL.String()

		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = strings.Split(forwardedFor, ",")[0]
			} else {
				ip = strings.Split(r.RemoteAddr, ":")[0]
			}
		}

		log.Printf("%s UA: %q Referer: %q IP: %s Method: %s URL: %s",
			start.Format(time.RFC3339),
			r.UserAgent(),
			r.Referer(),
			ip,
			r.Method,
			URL,
		)
		next.ServeHTTP(w, r)
		log.Printf("took Duration: %s, URL: %s",
			time.Since(start).String(),
			URL,
		)
	})
}

func (api *API) attributionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("API-Name", "Go-quote")
		w.Header().Set("Attribution", "If this API has helped you, please star it on GitHub: https://github.com/Attumm/go_quote")
		w.Header().Set("Support", "For issues or questions visit: https://github.com/Attumm/go_quote/issues")
		w.Header().Set("Docs", "For swagger docs visit: <url>/docs/")

		next.ServeHTTP(w, r)
	})
}

func (api *API) SetupMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if api.EnableLogging {
			next = api.logMiddleware(next)
		}
		if api.PermissiveCORS {
			next = api.corsMiddleware(next)
		}
		next = api.attributionMiddleware(next)
		return next
	}
}

func (api *API) SetupRoutes(mux *http.ServeMux) {

	mux.HandleFunc("/quotes", api.ListQuotesHandler)
	mux.HandleFunc("/quotes/", api.QuoteHandler)

	mux.HandleFunc("/tags", api.ListTagsHandler)
	mux.HandleFunc("/tags/", api.TagQuotesHandler)

	mux.HandleFunc("/authors", api.ListAuthorsHandler)
	mux.HandleFunc("/authors/", api.AuthorQuotesHandler)

	mux.HandleFunc("/random-quote", api.QuoteHandler)

	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		PrintMemUsage()
		fmt.Fprintf(w, "Debug information printed to console")
	})
	mux.HandleFunc("/favicon.ico", api.faviconHandler)
	mux.HandleFunc("/examples/", api.HandleFormatDocs)

	mux.HandleFunc("/", api.QuoteHandler)

	if api.Swagger {
		opts := middleware.SwaggerUIOpts{SpecURL: "/swagger.json"}
		sh := middleware.SwaggerUI(opts, nil)
		mux.Handle("/docs/", sh)
		mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(SwaggerSpec))
		})
	}
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

type PaginatedQuotesResponse struct {
	Quotes     []ResponseQuote `json:"quotes"`
	Pagination Pagination      `json:"pagination"`
}

type AuthorResponse struct {
	Name        string `json:"name"`
	AuthorID    string `json:"author_id"`
	TotalQuotes int    `json:"total_quotes"`
}

type PaginatedAuthorResponse struct {
	Author      string          `json:"author"`
	AuthorID    string          `json:"author_id"`
	TotalQuotes int             `json:"total_quotes"`
	Quotes      []ResponseQuote `json:"quotes"`
	Pagination  Pagination      `json:"pagination"`
}

type PaginatedAuthorsResponse struct {
	Authors    []AuthorResponse `json:"authors"`
	Pagination Pagination       `json:"pagination"`
}

type TagResponse struct {
	Name        string `json:"name"`
	TagID       string `json:"tag_id"`
	TotalQuotes int    `json:"total_quotes"`
}

type PaginatedTagsResponse struct {
	Tags       []TagResponse `json:"tags"`
	Pagination Pagination    `json:"pagination"`
}

func (api *API) TagQuotesHandler(w http.ResponseWriter, r *http.Request) {
	tagName := r.URL.Path[len("/tags/"):]

	quoteIDs, exists := api.Tags.NameToQuotes[tagName]
	if !exists {
		returnError(w, getOutputFormat(r), http.StatusNotFound, "Tag not found", "Given tag does not exist")
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

	response := PaginatedQuotesResponse{
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

	response := PaginatedTagsResponse{
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

	response := PaginatedAuthorsResponse{
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
		returnError(w, getOutputFormat(r), http.StatusNotFound, "Author not found", "Given author does not exist")
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

	response := PaginatedAuthorResponse{
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
		fmt.Println("Invalid data type provided")
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
		QuoteURL: fmt.Sprintf("%s/quotes/%d", baseURL, quoteID),
	}
}

func (api *API) ListQuotesHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestDataList(r, api, QuotesTypeRequest)
	api.formatStreamingResponse(w, requestData)
}

func (api *API) QuoteHandler(w http.ResponseWriter, r *http.Request) {
	requestData := createRequestData(r, api)
	if len(api.Quotes) == 0 {
		returnError(w, getOutputFormat(r), http.StatusNotFound, "No quotes available", "The quote database is empty")
		return
	}

	quoteID := rand.Intn(len(api.Quotes))
	var err error
	if strings.Contains(r.URL.Path, "quotes/") {
		quoteID, err = getID(r.URL.Path)
		if err != nil || quoteID < 0 || quoteID >= len(api.Quotes) {
			returnError(w, getOutputFormat(r), http.StatusNotFound, "Quote not found", fmt.Sprintf("Invalid quote ID"))
			return
		}
	}
	quote := api.Quotes[quoteID].CreateResponseQuote(quoteID)

	responseInfo := getResponseInfo(r, quoteID, requestData)
	api.formatResponseQuote(w, quote, responseInfo)
}

const faviconIcon = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <rect width="32" height="32" fill="#4A5568" rx="6"/>
  <path d="M11 22V17.3C11 14.4 12 11.7 15 10L16.5 11.5C14.5 13 13.8 14.8 13.7 16.5H16V22H11Z" fill="#E2E8F0"/>
  <path d="M22 22V17.3C22 14.4 23 11.7 26 10L27.5 11.5C25.5 13 24.8 14.8 24.7 16.5H27V22H22Z" fill="#E2E8F0"/>
</svg>
`

func (api *API) faviconHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	w.Write([]byte(faviconIcon))
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
        <pre>GET /quotes/1?format=json</pre>
    </div>

    <div class="method">
        <h2>Method 2: Using Accept Header</h2>
        <p>Set the appropriate Accept header in your request:</p>
        <pre>curl -H "Accept: application/json" /quotes/1</pre>
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
                <a href="/quotes/1?format={{.Format}}" target="_blank">View {{.Format}} format</a>
            </div>

            <h4>Using format parameter:</h4>
            <pre>GET /quotes/1?format={{.Format}}</pre>

            <h4>Using Accept header:</h4>
            <pre>curl -H "Accept: {{.ContentType}}" /quotes/1</pre>

            <div class="example-preview">
                <div class="preview-tabs">
                    <button class="preview-tab active" onclick="showTab(this, 'preview-{{.Format}}')">Preview</button>
                    <button class="preview-tab" onclick="showTab(this, 'source-{{.Format}}')">Source</button>
                </div>
                <div id="preview-{{.Format}}" class="preview-content">
                    {{if .IsAudio}}
                        <audio controls>
                            <source src="/quotes/1?format={{.Format}}" type="{{.ContentType}}">
                            Your browser does not support the audio element.
                        </audio>
                    {{else}}
                        <iframe src="/quotes/1?format={{.Format}}"></iframe>
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

type ErrorResponse struct {
	Status  int    `json:"status" xml:"status"`
	Message string `json:"message" xml:"message"`
	Error   string `json:"error" xml:"error"`
}

func returnError(w http.ResponseWriter, format string, status int, message string, err string) {
	errorResponse := ErrorResponse{
		Status:  status,
		Message: message,
		Error:   err,
	}

	w.WriteHeader(status)
	formatErrorResponse(w, errorResponse, format)
}

func formatErrorResponse(w http.ResponseWriter, errorResponse ErrorResponse, format string) {
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errorResponse)
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(xml.Header))
		xml.NewEncoder(w).Encode(errorResponse)
	case "html":
		w.Header().Set("Content-Type", "text/html")
		htmlTemplate := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Error: {{.Status}}</title>
        </head>
        <body>
            <h1>Error {{.Status}}</h1>
            <p><strong>Message:</strong> {{.Message}}</p>
            <p><strong>Error:</strong> {{.Error}}</p>
        </body>
        </html>`
		tmpl, _ := template.New("error").Parse(htmlTemplate)
		tmpl.Execute(w, errorResponse)
	case "text":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Error %d\nMessage: %s\nError: %s\n", errorResponse.Status, errorResponse.Message, errorResponse.Error)
	case "yaml":
		w.Header().Set("Content-Type", "application/x-yaml")
		yaml.NewEncoder(w).Encode(errorResponse)
	case "markdown":
		w.Header().Set("Content-Type", "text/markdown")
		markdownTemplate := `
# Error {{.Status}}

**Message:** {{.Message}}

**Error:** {{.Error}}
`
		tmpl, _ := template.New("error").Parse(markdownTemplate)
		tmpl.Execute(w, errorResponse)
	default:
		// If an unknown format is requested, default to JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errorResponse)
	}
}
