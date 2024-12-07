package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var OutputFormats = map[string]string{
	"xml":        "application/xml",
	"html":       "text/html",
	"json":       "application/json",
	"text":       "text/plain",
	"markdown":   "text/markdown",
	"yaml":       "application/yaml",
	"csv":        "text/csv",
	"rss":        "application/rss+xml",
	"atom":       "application/atom+xml",
	"oembed":     "application/json+oembed",
	"oembed.xml": "text/xml+oembed",
	"embed":      "text/html",
	"embed.js":   "application/javascript",
	"anker":      "application/json",
	"json1":      "application/json",
	"json2":      "application/json",
	"json3":      "application/json",
	"json4":      "application/json",
	"json5":      "application/json",
	"json6":      "application/json",
	"json7":      "application/json",
	"json8":      "application/json",
}

type OEmbedResponse struct {
	Type         string `json:"type"`
	Version      string `json:"version"`
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url,omitempty"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	CacheAge     int    `json:"cache_age"`
	HTML         string `json:"html"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

func (oe *OEmbedResponse) Create(quote ResponseQuote, baseURL string) {
	oe.Type = "rich"
	oe.Version = "1.0"
	oe.Title = "Quote by " + quote.Author
	oe.AuthorName = quote.Author
	oe.ProviderName = "Quotes API"
	oe.ProviderURL = baseURL
	oe.CacheAge = 3600
	oe.HTML = quoteToEmbeddedHTML(quote)
	oe.Width = 500
	oe.Height = 200
}

type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	XMLNS   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
	Updated string     `xml:"updated"`
	Author  AtomAuthor `xml:"author"`
	ID      string     `xml:"id"`
	Entry   AtomEntry  `xml:"entry"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomEntry struct {
	Title     string     `xml:"title"`
	Link      AtomLink   `xml:"link"`
	ID        string     `xml:"id"`
	Updated   string     `xml:"updated"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Author    AtomAuthor `xml:"author"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func (af *AtomFeed) Create(quote ResponseQuote, baseURL, quoteURL string) {
	now := time.Now().Format(time.RFC3339)

	af.XMLNS = "http://www.w3.org/2005/Atom"
	af.Title = "Random Quote"
	af.Updated = now
	af.Author = AtomAuthor{Name: "Quotes API"}
	af.ID = quoteURL
	af.Link = struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	}{
		Href: baseURL,
		Rel:  "self",
	}
	af.Entry = AtomEntry{
		Title:     "Quote by " + quote.Author,
		Updated:   now,
		Published: now,
		ID:        quoteURL,
		Content:   quote.Text,
		Author:    AtomAuthor{Name: quote.Author},
		Link:      AtomLink{Href: quoteURL},
	}
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
func quotesToHTML(response struct {
	Quotes     []ResponseQuote `json:"quotes"`
	Pagination Pagination      `json:"pagination"`
}, tagName string) string {
	var htmlBuilder strings.Builder

	htmlBuilder.WriteString(`
    <div class="quotes-container">
        <style>
            .quotes-container {
                display: flex;
                flex-wrap: wrap;
                justify-content: space-between;
                max-width: 1200px;
                margin: 0 auto;
                padding: 20px;
            }
            .pagination {
                width: 100%;
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-top: 20px;
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', sans-serif;
            }
            .pagination-info {
                color: #657786;
                font-size: 14px;
            }
            .pagination-link {
                background-color: #f1f3f5;
                color: #657786;
                padding: 8px 12px;
                border-radius: 4px;
                text-decoration: none;
                font-size: 14px;
            }
            .pagination-link:hover {
                background-color: #e1e8ed;
                color: #1da1f2;
            }
            .tag-header {
                width: 100%;
                background-color: #f1f3f5;
                padding: 20px;
                margin-bottom: 20px;
                border-radius: 4px;
            }
            .tag-header h1 {
                margin: 0;
                color: #14171a;
                font-size: 24px;
            }
            .tag-header p {
                margin: 10px 0 0;
                color: #657786;
            }
        </style>
    `)

	if tagName != "" {
		htmlBuilder.WriteString(fmt.Sprintf(`
        <div class="tag-header">
            <h1>Quotes tagged with "%s"</h1>
            <p>Found %d quotes</p>
        </div>
        `, tagName, response.Pagination.Total))
	}

	for _, quote := range response.Quotes {
		quoteHTML := quoteToHTML(quote, tagName)
		htmlBuilder.WriteString(quoteHTML)
	}

	htmlBuilder.WriteString(`
        <div class="pagination">
            <span class="pagination-info">
                Page ` + fmt.Sprintf("%d", response.Pagination.Page) + ` of ` + fmt.Sprintf("%d", response.Pagination.Pages) + `
            </span>
            <div>
                <a href="?page=` + fmt.Sprintf("%d", response.Pagination.Page-1) + `" class="pagination-link">Previous</a>
    `)

	if response.Pagination.Next != "" {
		htmlBuilder.WriteString(`<a href="` + response.Pagination.Next + `" class="pagination-link">Next</a>`)
	}

	htmlBuilder.WriteString(`
            </div>
        </div>
    </div>`)

	return htmlBuilder.String()
}

func quoteToHTML(quote ResponseQuote, highlightedTag string) string {
	var tagHTML strings.Builder
	for _, tag := range quote.Tags {
		cleanTag := strings.TrimSpace(tag)
		encodedTag := url.QueryEscape(cleanTag)
		tagClass := "tag"
		if cleanTag == highlightedTag {
			tagClass += " highlighted-tag"
		}
		tagHTML.WriteString(fmt.Sprintf(`<a href="/tags/%s" class="%s">%s</a>`, encodedTag, tagClass, cleanTag))
	}

	authorLink := fmt.Sprintf(`<a href="/authors/%s" class="author">— %s</a>`,
		strings.TrimSpace(quote.AuthorID),
		strings.TrimSpace(quote.Author))

	return fmt.Sprintf(`
    <div class="quote-container">
        <p class="quote-text">%s</p>
        %s
        <div class="tags">%s</div>
        <div class="footer">
            <span class="quote-id">Quote ID: %d</span>
            <button class="next-button" onclick="location.href='/random-quote'">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M23 4v6h-6"/>
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                </svg>
            </button>
        </div>
    </div>
    <style>
        html {
            -webkit-text-size-adjust: 100%%;
            -ms-text-size-adjust: 100%%;
        }
        body {
            text-size-adjust: none;
            -webkit-text-size-adjust: none;
            -moz-text-size-adjust: none;
            -ms-text-size-adjust: none;
        }
        .quote-container {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', sans-serif;
            max-width: 500px;
            margin: 20px auto;
            padding: 20px;
            background-color: #ffffff;
            border: 1px solid #e1e8ed;
            border-radius: 4px;
        }
        .quote-text {
            font-size: 18px !important;
            line-height: 1.5;
            color: #333333;
            margin-bottom: 15px;
            font-style: italic;
            max-width: 100%%;
            overflow-wrap: break-word;
            word-wrap: break-word;
        }
        .author {
            font-size: 16px !important;
            color: #657786;
            margin-bottom: 15px;
            text-decoration: none;
        }
        .author:hover {
            color: #1da1f2;
        }
        .tags {
            margin-bottom: 15px;
        }
        .tag {
            display: inline-block;
            background-color: #f1f3f5;
            color: #657786;
            padding: 4px 8px;
            border-radius: 16px;
            font-size: 14px !important;
            margin-right: 5px;
            margin-bottom: 5px;
            text-decoration: none;
        }
        .tag:hover {
            background-color: #e1e8ed;
            color: #1da1f2;
        }
        .highlighted-tag {
            background-color: #1da1f2 !important;
            color: white !important;
        }
        .highlighted-tag:hover {
            background-color: #1a91da !important;
        }
        .footer {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .quote-id {
            color: #aab8c2;
            font-size: 12px !important;
        }
        .next-button {
            background-color: transparent;
            color: #657786;
            border: none;
            padding: 4px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .next-button:hover {
            color: #1da1f2;
        }
        @media screen and (max-width: 600px) {
            .quote-text {
                font-size: 16px !important;
            }
            .author {
                font-size: 14px !important;
            }
            .tag {
                font-size: 12px !important;
            }
        }
    </style>
    `, quote.Text, authorLink, tagHTML.String(), quote.ID)
}

type XMLQuote struct {
	XMLName xml.Name `xml:"response"`
	ID      int      `xml:"id"`
	Text    string   `xml:"text"`
	Author  string   `xml:"author"`
	Tags    []string `xml:"tags>tag"`
}

func quoteToMarkdown(quote ResponseQuote) string {
	tags := strings.Join(quote.Tags, " #")
	if len(tags) > 0 {
		tags = "#" + tags
	}
	return fmt.Sprintf("> %s\n\n— %s\n\nQuote ID: %d\n\n%s", quote.Text, quote.Author, quote.ID, tags)
}

func quoteToRSS(quote ResponseQuote, quoteUrl string) string {
	pubDate := time.Now().Format(time.RFC1123Z)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>Random Quote</title>
  <link>/quotes/%s</link>
  <description>Random Quote of the Moment</description>
  <item>
    <title>Quote by %s</title>
    <description>%s</description>
    <author>%s</author>
    <pubDate>%s</pubDate>
    <guid>%v</guid>
  </item>
</channel>
</rss>`, quoteUrl, quote.Author, quote.Text, quote.Author, pubDate, quoteUrl)
}

func quoteToYAML(quote ResponseQuote) string {
	return fmt.Sprintf(`quote:
  id: %d
  text: "%s"
  author: "%s"
  tags:
    - %s
`, quote.ID, quote.Text, quote.Author, strings.Join(quote.Tags, "\n    - "))
}
func quotesToYAML(quotes []ResponseQuote) string {
	var buf strings.Builder
	buf.WriteString("quotes:\n")
	for _, quote := range quotes {
		buf.WriteString(fmt.Sprintf("  - id: %d\n", quote.ID))
		buf.WriteString(fmt.Sprintf("    text: \"%s\"\n", strings.ReplaceAll(quote.Text, "\"", "\\\"")))
		buf.WriteString(fmt.Sprintf("    author: \"%s\"\n", strings.ReplaceAll(quote.Author, "\"", "\\\"")))
		buf.WriteString("    tags:\n")
		for _, tag := range quote.Tags {
			buf.WriteString(fmt.Sprintf("      - %s\n", tag))
		}
	}
	return buf.String()
}

func quoteToCSV(quote ResponseQuote) string {
	var sb strings.Builder
	sb.WriteString(`"ID","Text","Author","Tags"`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`"%d","%s","%s","%s"`,
		quote.ID,
		strings.ReplaceAll(quote.Text, `"`, `""`),
		strings.ReplaceAll(quote.Author, `"`, `""`),
		strings.Join(quote.Tags, "|")))
	return sb.String()
}

func quotesToCSV(quotes ResponseQuotes) string {
	var buf bytes.Buffer
	buf.WriteString(`"ID","Text","Author","Tags"`)
	buf.WriteString("\n")

	for _, quote := range quotes {
		buf.WriteString(fmt.Sprintf(`"%d","%s","%s","%s"`,
			quote.ID,
			strings.ReplaceAll(quote.Text, `"`, `""`),
			strings.ReplaceAll(quote.Author, `"`, `""`),
			strings.Join(quote.Tags, "|")))
		buf.WriteString("\n")
	}

	return buf.String()
}

type ResponseInfo struct {
	QuoteID  int
	BaseURL  string
	QuoteURL string
	Format   string
}

func getResponseInfo(r *http.Request, quoteID int) ResponseInfo {
	baseURL := fmt.Sprintf("%s://%s", scheme(r), r.Host)
	return ResponseInfo{
		QuoteID:  quoteID,
		BaseURL:  baseURL,
		QuoteURL: fmt.Sprintf("%s/quote/%d", baseURL, quoteID),
		Format:   getOutputFormat(r),
	}
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

func (api *API) formatResponseQuote(w http.ResponseWriter, quote ResponseQuote, responseInfo ResponseInfo) {
	switch responseInfo.Format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quote)

	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		xmlQuote := XMLQuote{
			Text:   quote.Text,
			Author: quote.Author,
			Tags:   quote.Tags,
			ID:     quote.ID,
		}
		xml.NewEncoder(w).Encode(xmlQuote)

	case "html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, quoteToHTML(quote, ""))

	case "text":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Quote: %s\nAuthor: %s\nTags: %v\nID: %d\n",
			quote.Text, quote.Author, strings.Join(quote.Tags, ", "), quote.ID)

	case "markdown":
		w.Header().Set("Content-Type", "text/markdown")
		fmt.Fprint(w, quoteToMarkdown(quote))

	case "yaml":
		w.Header().Set("Content-Type", "application/yaml")
		fmt.Fprint(w, quoteToYAML(quote))

	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		fmt.Fprint(w, quoteToCSV(quote))

	case "rss":
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, quoteToRSS(quote, responseInfo.QuoteURL))

	case "atom":
		w.Header().Set("Content-Type", "application/atom+xml")
		feed := &AtomFeed{}
		feed.Create(quote, responseInfo.BaseURL, responseInfo.QuoteURL)
		xml.NewEncoder(w).Encode(feed)

	case "oembed", "oembed.json":
		w.Header().Set("Content-Type", "application/json+oembed")
		response := &OEmbedResponse{}
		response.Create(quote, responseInfo.BaseURL)
		json.NewEncoder(w).Encode(response)

	case "oembed.xml":
		w.Header().Set("Content-Type", "text/xml+oembed")
		response := &OEmbedResponse{}
		response.Create(quote, responseInfo.BaseURL)
		xml.NewEncoder(w).Encode(response)

	case "embed":
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, quoteToEmbeddedHTML(quote))

	case "embed.js":
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, quoteToEmbeddedJS(quote))

	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quote)
	}
}

func quoteToEmbeddedHTML(quote ResponseQuote) string {
	return fmt.Sprintf(`
<blockquote class="quote-embed" 
    style="font-family: Arial, sans-serif; max-width: 500px; margin: 20px auto; padding: 20px;
           border-left: 5px solid #eee; background-color: #f9f9f9;">
    <p style="font-size: 18px; line-height: 1.4; margin-bottom: 10px;">"%s"</p>
    <footer style="color: #666;">
        — <cite>%s</cite>
    </footer>
    <div style="margin-top: 10px; font-size: 14px; color: #888;">
        ID: %d • %s
    </div>
</blockquote>`, quote.Text, quote.Author, quote.ID, strings.Join(quote.Tags, " • "))
}

func quoteToEmbeddedJS(quote ResponseQuote) string {
	return fmt.Sprintf(`
(function() {
    var quote = %s;
    var container = document.currentScript.parentElement;
    container.innerHTML = %s;
})();`,
		mustJSONString(quote),
		mustJSONString(quoteToEmbeddedHTML(quote)))
}

func mustJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
