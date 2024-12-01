package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func LoadQuotes(filename, storageType string) (Quotes, error) {
	switch storageType {
	case "csv":
		return LoadQuotesFromCSV(filename)
	case "bytes":
		return LoadAsBytes(filename)
	case "bytesz":
		return LoadAsBytesCompressed(filename)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// SaveQuotes saves quotes based on the storage type
func SaveQuotes(quotes Quotes, filename, storageType string) error {
	switch storageType {
	case "csv":
		return SaveQuotesToCSV(quotes, filename)
	case "bytes":
		_, err := SaveAsBytes(quotes, filename)
		return err
	case "bytesz":
		_, err := SaveAsBytesCompressed(quotes, filename)
		return err
	default:
		return fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

func ConvertQuotes(inputFilename, inputStorageType, outputFilename, outputStorageType string) error {
	quotes, err := LoadQuotes(inputFilename, inputStorageType)
	if err != nil {
		return fmt.Errorf("error loading quotes: %v", err)
	}

	err = SaveQuotes(quotes, outputFilename, outputStorageType)
	if err != nil {
		return fmt.Errorf("error saving quotes: %v", err)
	}

	return nil
}

// LoadQuotesFromCSV loads quotes from a CSV file and returns a Quotes slice
func LoadQuotesFromCSV(filename string) (Quotes, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and discard the header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("error reading CSV header: %v", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	quotes := make(Quotes, 0, len(records))
	for _, record := range records {
		if len(record) != 3 {
			fmt.Printf("Skipping invalid record: %v\n", record)
			continue
		}

		quote := Quote{
			Text:   record[0],
			Author: record[1],
			Tags:   strings.Split(record[2], ", "),
		}
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// SaveQuotesToCSV saves quotes to a CSV file
func SaveQuotesToCSV(quotes Quotes, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"quote", "author", "category"})
	if err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	// Write quotes
	for _, quote := range quotes {
		err := writer.Write([]string{
			quote.Text,
			quote.Author,
			strings.Join(quote.Tags, ", "),
		})
		if err != nil {
			return fmt.Errorf("error writing quote to CSV: %v", err)
		}
	}

	return nil
}

// SaveAsBytes encodes Quotes to bytes and saves to a file
func SaveAsBytes(quotes Quotes, filename string) (int64, error) {
	data := EncodeQuotes(quotes)
	err := WriteToFile(data, filename)
	if err != nil {
		return 0, err
	}
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// SaveAsBytesCompressed encodes Quotes to bytes, compresses, and saves to a file
func SaveAsBytesCompressed(quotes Quotes, filename string) (int64, error) {
	data := EncodeQuotes(quotes)
	data = Compress(data)
	err := WriteToFile(data, filename)
	if err != nil {
		return 0, err
	}
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// EncodeQuotes encodes Quotes to bytes using gob
func EncodeQuotes(quotes Quotes) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(quotes)
	if err != nil {
		fmt.Println("Error encoding:", err)
	}
	return buf.Bytes()
}

// Compress compresses byte slice using gzip
func Compress(s []byte) []byte {
	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	zipped.Write(s)
	zipped.Close()
	return zipbuf.Bytes()
}

// Decompress decompresses byte slice using gzip
func Decompress(s []byte) []byte {
	reader, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println("Unable to decompress:", err)
	}
	reader.Close()
	return data
}

// WriteToFile writes byte slice to a file
func WriteToFile(s []byte, filename string) error {
	return ioutil.WriteFile(filename, s, 0644)
}

// ReadFromFile reads byte slice from a file
func ReadFromFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// LoadAsBytes loads Quotes from a file stored as bytes
func LoadAsBytes(filename string) (Quotes, error) {
	data, err := ReadFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}
	return DecodeToQuotes(data)
}

// LoadAsBytesCompressed loads Quotes from a file stored as compressed bytes
func LoadAsBytesCompressed(filename string) (Quotes, error) {
	data, err := ReadFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}
	data = Decompress(data)
	return DecodeToQuotes(data)
}

// DecodeToQuotes decodes byte slice to Quotes using gob
func DecodeToQuotes(s []byte) (Quotes, error) {
	quotes := make(Quotes, 0)
	decoder := gob.NewDecoder(bytes.NewReader(s))
	err := decoder.Decode(&quotes)
	if err != nil {
		return nil, fmt.Errorf("unable to decode to Quotes: %v", err)
	}
	return quotes, nil
}
