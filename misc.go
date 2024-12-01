package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func getConvertedFilename(originalFilename, convertStorage, outputDir string) string {
	filename := filepath.Base(originalFilename)
	extension := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(extension)]

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Printf("Error creating output directory: %v", err)
		outputDir = "."
	}

	return filepath.Join(outputDir, nameWithoutExt+"."+convertStorage)
}

func getID(path string) (int, error) {
	if strings.Contains(path, "/quote/") {
		parts := strings.Split(path, "/quote/")
		if len(parts) != 2 {
			return -1, fmt.Errorf("invalid path format")
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			return -1, err
		}
		return id, nil
	}
	return -1, nil
}
