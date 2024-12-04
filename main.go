package main

import (
	"fmt"
	"github.com/Attumm/settingo/settingo"
	"log"
	"net/http"
	"runtime"
)

type Config struct {
	Filename        string `settingo:"Path of the filename"`
	Storage         string `settingo:"Storage type"`
	Convert         bool   `settingo:"Convert mode, will convert data into Convert Storage type"`
	ConvertStorage  string `settingo:"Storage type to convert to"`
	OutputDir       string `settingo:"Directory to store converted files"`
	Port            string `settingo:"Port for the API server"`
	Host            string `settingo:"Host for the API server"`
	DefaultPageSize int    `settingo:"Page size to use for the API server"`
	MaxPageSize     int    `settingo:"Maximum quotes for the API server"`
}

func main() {
	config := &Config{
		Filename:        "data/quotes.csv",
		Storage:         "csv",
		Convert:         false,
		ConvertStorage:  "bytesz",
		OutputDir:       "data",
		Port:            "8000",
		Host:            "0.0.0.0",
		DefaultPageSize: 10,
		MaxPageSize:     1000000,
	}

	settingo.ParseTo(config)

	PrintMemUsage()

	runtime.GC()
	quotes, err := LoadQuotes(config.Filename, config.Storage)
	if err != nil {
		log.Fatalf("Error loading quotes: %v", err)
	}
	fmt.Printf("Loaded %d quotes from %s\n", len(quotes), config.Filename)
	runtime.GC()
	/*
		if config.Convert {
			outputFilename := getConvertedFilename(config.Filename, config.ConvertStorage, config.OutputDir)
			err := ConvertQuotes(quotes, outputFilename, config.ConvertStorage)

			if err != nil {
				log.Fatalf("Error converting quotes: %v", err)
			}
			fmt.Printf("Converted quotes to %s and saved as %s\n", config.ConvertStorage, outputFilename)
		}

	*/
	PrintMemUsage()
	runtime.GC()
	fmt.Printf("Total quotes processed: %d\n", len(quotes))
	PrintMemUsage()

	authorIndex := BuildAuthorIndex(quotes)
	tagIndex := BuildTagIndex(quotes)

	api := &API{
		Quotes:          quotes,
		Authors:         authorIndex,
		Tags:            tagIndex,
		DefaultPageSize: config.DefaultPageSize,
		MaxPageSize:     config.MaxPageSize,
	}

	api.SetupRoutes()

	fmt.Printf("Starting server on port %s...\n", config.Port)
	if err := http.ListenAndServe(config.Host+":"+config.Port, nil); err != nil {
		log.Fatal(err)
	}
}
