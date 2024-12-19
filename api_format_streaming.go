package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func streamQuotesJSON(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	setPaginationHeaders(w, RequestDataList.Pagination)

	var writer io.Writer
	var bufWriter *bufio.Writer

	approximateSize := RequestDataList.Total*200 + 100

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		bufWriter = bufio.NewWriterSize(gw, approximateSize)
		defer bufWriter.Flush()
		defer gw.Close()
		writer = bufWriter
	} else {
		bufWriter = bufio.NewWriterSize(w, approximateSize)
		defer bufWriter.Flush()
		writer = bufWriter
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Pre-allocate the JSON encoder
	enc := json.NewEncoder(writer)

	if _, err := writer.Write([]byte(`{"quotes":[`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex-1; i++ {
		if err := enc.Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := writer.Write([]byte(",")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Handle the last item without a trailing comma
	if RequestDataList.EndIndex > RequestDataList.StartIndex {
		if err := enc.Encode(api.Quotes[RequestDataList.EndIndex-1].CreateResponseQuote(RequestDataList.EndIndex - 1)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if _, err := writer.Write([]byte(`],"pagination":`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := enc.Encode(RequestDataList.Pagination); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write([]byte("}")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bufWriter.Flush(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func streamQuotesCSV(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=quotes.csv")
	setPaginationHeaders(w, RequestDataList.Pagination)

	var writer io.Writer
	var bufWriter *bufio.Writer
	const bufferSize = 64 * 1024 // 64KB buffer, adjust as needed

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		bufWriter = bufio.NewWriterSize(gw, bufferSize)
		writer = bufWriter
	} else {
		bufWriter = bufio.NewWriterSize(w, bufferSize)
		writer = bufWriter
	}

	defer bufWriter.Flush()

	w.WriteHeader(http.StatusOK)

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{"ID", "Text", "Author", "AuthorID", "Tags"}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, "Failed to write header row: "+err.Error(), http.StatusInternalServerError)
		return
	}

	row := make([]string, 5)
	fmt.Println(RequestDataList)
	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		quote := api.Quotes[i].CreateResponseQuote(i)

		row[0] = strconv.Itoa(quote.ID)
		row[1] = quote.Text
		row[2] = quote.Author
		row[3] = quote.AuthorID
		row[4] = strings.Join(quote.Tags, "|")

		if err := csvWriter.Write(row); err != nil {
			http.Error(w, "Failed to write data row: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func streamQuotesXML(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/xml")

	var writer io.Writer
	var bufWriter *bufio.Writer
	const bufferSize = 64 * 1024 // 64KB buffer, adjust as needed

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		bufWriter = bufio.NewWriterSize(gw, bufferSize)
		writer = bufWriter
	} else {
		bufWriter = bufio.NewWriterSize(w, bufferSize)
		writer = bufWriter
	}

	defer bufWriter.Flush()

	w.WriteHeader(http.StatusOK)

	xmlEncoder := xml.NewEncoder(writer)
	defer xmlEncoder.Flush()

	// Write XML header
	io.WriteString(writer, xml.Header)
	io.WriteString(writer, "<quotes>")

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		quote := api.Quotes[i].CreateResponseQuote(i)
		xmlQuote := XMLQuote{
			Text:   quote.Text,
			Author: quote.Author,
			Tags:   quote.Tags,
			ID:     quote.ID,
		}

		if err := xmlEncoder.Encode(xmlQuote); err != nil {
			http.Error(w, "Failed to encode XML: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	io.WriteString(writer, "</quotes>")
}

type XMLQuotes struct {
	XMLName xml.Name `xml:"quote"`
	Text    string   `xml:"text"`
	Author  string   `xml:"author"`
	Tags    []string `xml:"tags>tag"`
	ID      int      `xml:"id,attr"`
}

func streamQuotesYAML(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=quotes.yaml")

	setPaginationHeaders(w, RequestDataList.Pagination)

	var writer io.Writer
	var bufWriter *bufio.Writer
	const bufferSize = 64 * 1024 // 64KB buffer, adjust as needed

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		bufWriter = bufio.NewWriterSize(gw, bufferSize)
		writer = bufWriter
	} else {
		bufWriter = bufio.NewWriterSize(w, bufferSize)
		writer = bufWriter
	}

	defer bufWriter.Flush()

	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprintln(writer, "quotes:"); err != nil {
		http.Error(w, "Failed to write YAML header: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		quote := api.Quotes[i].CreateResponseQuote(i)

		if _, err := fmt.Fprintf(writer, "  - id: %d\n", quote.ID); err != nil {
			http.Error(w, "Failed to write quote ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(writer, "    text: \"%s\"\n", strings.ReplaceAll(quote.Text, "\"", "\\\"")); err != nil {
			http.Error(w, "Failed to write quote text: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(writer, "    author: \"%s\"\n", strings.ReplaceAll(quote.Author, "\"", "\\\"")); err != nil {
			http.Error(w, "Failed to write quote author: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintln(writer, "    tags:"); err != nil {
			http.Error(w, "Failed to write tags header: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, tag := range quote.Tags {
			if _, err := fmt.Fprintf(writer, "      - %s\n", tag); err != nil {
				http.Error(w, "Failed to write tag: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func (api *API) formatStreamingResponse(w http.ResponseWriter, RequestDataList *RequestDataList) {
	setPaginationHeaders(w, RequestDataList.Pagination)
	switch RequestDataList.Format {
	case "json":
		streamQuotesJSON(w, api, RequestDataList)
	case "csv":
		streamQuotesCSV(w, api, RequestDataList)
	case "yaml":
		streamQuotesYAML(w, api, RequestDataList)
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}
