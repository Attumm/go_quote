package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Writer interface {
	io.Writer
	Flush() error
}

func getWriter(w http.ResponseWriter, useGzip bool) (io.Writer, func() error, error) {
	var writer io.Writer
	var cleanup func() error

	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		writer = gw
		cleanup = gw.Close
	} else {
		bw := bufio.NewWriter(w)
		writer = bw
		cleanup = bw.Flush
	}

	return writer, cleanup, nil
}

func streamQuotesJSON(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	if RequestDataList.Total == 1 {
		normalQuotesJSON(w, api, RequestDataList)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var writer Writer
	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		writer = gw
	} else {
		bw := bufio.NewWriter(w)
		defer bw.Flush()
		writer = bw
	}

	if _, err := writer.Write([]byte(`{"quotes":[`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isFirst := true
	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		if !isFirst {
			if _, err := writer.Write([]byte(",")); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		isFirst = false

		if err := json.NewEncoder(writer).Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println("ERROR", err, api.Quotes[i])
			return
		}
	}

	if _, err := writer.Write([]byte(`],"pagination":`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(writer).Encode(RequestDataList.Pagination); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := writer.Write([]byte("}")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := writer.Flush(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func normalQuotesJSON(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Quotes     []ResponseQuote `json:"quotes"`
		Pagination Pagination      `json:"pagination"`
	}{
		Quotes:     make([]ResponseQuote, 0, RequestDataList.EndIndex-RequestDataList.StartIndex),
		Pagination: RequestDataList.Pagination,
	}

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		response.Quotes = append(response.Quotes, api.Quotes[i].CreateResponseQuote(i))
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if false && RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gw.Close()
		if _, err := gw.Write(buf.Bytes()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if _, err := w.Write(buf.Bytes()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func streamQuotesJSONMEM(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	bw := bufio.NewWriter(w)
	bw.Write([]byte(`{"quotes":[`))

	isFirst := true
	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		if !isFirst {
			bw.Write([]byte(","))
		} else {
			isFirst = false
		}

		if err := json.NewEncoder(bw).Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	bw.Write([]byte(`],"pagination":`))
	if err := json.NewEncoder(bw).Encode(RequestDataList.Pagination); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bw.Write([]byte("}"))
	bw.Flush()
}

func streamQuotesJSONMEM2(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Use a buffered writer for efficient output
	bw := bufio.NewWriter(w)

	// Start the JSON response
	bw.WriteString(`{"quotes":[`)

	// Buffer to collect quotes
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		if i > RequestDataList.StartIndex {
			bw.WriteString(",")
		}

		// Encode the quote directly into the buffer
		if err := encoder.Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the encoded quote without the trailing newline
		bw.Write(bytes.TrimSpace(buf.Bytes()))
		buf.Reset()
	}

	// Add pagination
	bw.WriteString(`],"pagination":`)
	if err := encoder.Encode(RequestDataList.Pagination); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Close the JSON response
	bw.WriteString("}")

	// Flush the buffered writer
	bw.Flush()
}

func streamQuotesJSONMEM3(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create a reusable JSON encoder
	encoder := json.NewEncoder(w)

	// Write the initial JSON fragment
	w.Write([]byte(`{"quotes":[`))

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		if i > RequestDataList.StartIndex {
			w.Write([]byte(`,`))
		}

		if err := encoder.Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Write the pagination and closing JSON fragment
	w.Write([]byte(`],"pagination":`))
	if err := encoder.Encode(RequestDataList.Pagination); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`}`))
}

func streamQuotesJSONV1(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	var writer io.Writer
	var bufWriter *bufio.Writer

	// Calculate the approximate size of the response
	approximateSize := RequestDataList.Total*200 + 100 // Assuming average quote size of 200 bytes + extra for pagination

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		bufWriter = bufio.NewWriterSize(gw, approximateSize)
		writer = bufWriter
	} else {
		bufWriter = bufio.NewWriterSize(w, approximateSize)
		writer = bufWriter
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Pre-allocate the JSON encoder
	enc := json.NewEncoder(writer)

	// Write the opening of the JSON object
	if _, err := writer.Write([]byte(`{"quotes":[`)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := RequestDataList.StartIndex; i < RequestDataList.EndIndex; i++ {
		if i > RequestDataList.StartIndex {
			if _, err := writer.Write([]byte(",")); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := enc.Encode(api.Quotes[i].CreateResponseQuote(i)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Write the closing of the quotes array and the pagination object
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

func streamQuotesJSONV3(w http.ResponseWriter, api *API, RequestDataList *RequestDataList) {
	var writer io.Writer
	var bufWriter *bufio.Writer

	// Calculate the approximate size of the response
	approximateSize := RequestDataList.Total*200 + 100

	if RequestDataList.Gzip {
		w.Header().Set("Content-Encoding", "gzip")
		gw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gw.Close()
		bufWriter = bufio.NewWriterSize(gw, approximateSize)
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

func (api *API) formatStreamingResponse(w http.ResponseWriter, RequestDataList *RequestDataList) {
	w.Header().Set("Current-Page", strconv.Itoa(RequestDataList.Pagination.Page))
	w.Header().Set("Page-Size", strconv.Itoa(RequestDataList.Pagination.PageSize))
	w.Header().Set("Total-Count", strconv.Itoa(RequestDataList.Pagination.Total))
	w.Header().Set("Total-Pages", strconv.Itoa(RequestDataList.Pagination.Pages))

	switch RequestDataList.Format {
	case "json":
		streamQuotesJSON(w, api, RequestDataList)
	case "anker":
		normalQuotesJSON(w, api, RequestDataList)
	case "json2":
		streamQuotesJSONMEM(w, api, RequestDataList)
	case "json3":
		streamQuotesJSONMEM2(w, api, RequestDataList)
	case "json4":
		streamQuotesJSONV1(w, api, RequestDataList)
	case "json5":
		streamQuotesJSONV3(w, api, RequestDataList)
	case "json6":
		streamQuotesJSONMEM3(w, api, RequestDataList)

	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}
