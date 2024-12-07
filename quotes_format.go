package main

import (
	"bytes"
	"fmt"
	"strings"
)

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
