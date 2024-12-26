package main

import (
	"fmt"
	"github.com/Attumm/settingo/settingo"
	"log"
	"net/http"
	"runtime"
	"time"
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
	MemoryDebugLog  bool   `settingo:"Enable periodic memory debug log"`
	EnableLogging   bool   `settingo:"Enable logging of requests"`
	PermissiveCORS  bool   `settingo:"Enable Permissive CORS"`
}

func logMemoryUsagePeriodically() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
		log.Printf("TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		log.Printf("Sys = %v MiB", bToMb(m.Sys))
		log.Printf("NumGC = %v\n", m.NumGC)
	}
}

func main() {
	config := &Config{
		Filename:        "data/quotes.bytesz",
		Storage:         "bytesz",
		Convert:         false,
		ConvertStorage:  "bytesz",
		OutputDir:       "data",
		Port:            "8000",
		Host:            "0.0.0.0",
		DefaultPageSize: 10,
		MaxPageSize:     1000000,
		MemoryDebugLog:  false,
		EnableLogging:   true,
		PermissiveCORS:  true,
	}

	settingo.ParseTo(config)

	fmt.Printf("Started go-quote with permissive cors: %v\n", config.PermissiveCORS)

	if config.MemoryDebugLog {
		go logMemoryUsagePeriodically()
	}

	runtime.GC()
	quotes, err := LoadQuotes(config.Filename, config.Storage)
	if err != nil {
		log.Fatalf("Error loading quotes: %v", err)
	}
	fmt.Printf("Loaded: %d quotes from %s\n", len(quotes), config.Filename)
	runtime.GC()

	fmt.Printf("Total quotes processed: %d\n", len(quotes))

	authorIndex := BuildAuthorIndex(quotes)
	tagIndex := BuildTagIndex(quotes)

	fmt.Printf("Created index for Authors: %d and Tags: %d\n", authorIndex.Len(), tagIndex.Len())

	api := &API{
		Quotes:          quotes,
		Authors:         authorIndex,
		Tags:            tagIndex,
		DefaultPageSize: config.DefaultPageSize,
		MaxPageSize:     config.MaxPageSize,
		Runtime:         runtime.GOOS,
		EnableLogging:   config.EnableLogging,
		PermissiveCORS:  config.PermissiveCORS,
	}

	mux := http.NewServeMux()
	middleware := api.SetupMiddleware()
	api.SetupRoutes(mux)

	fmt.Printf("Starting server on port %s:%s...\n", config.Host, config.Port)
	if err := http.ListenAndServe(config.Host+":"+config.Port, middleware(mux)); err != nil {
		log.Fatal(err)
	}
}
