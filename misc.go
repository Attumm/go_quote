package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
	if len(path) == 0 {
		return -1, nil
	}

	end := len(path)
	if path[end-1] == '/' {
		end--
	}

	start := end - 1
	for ; start >= 0; start-- {
		if path[start] == '/' {
			start++
			break
		}
	}

	if start < 0 {
		start = 0
	}

	idStr := path[start:end]
	if idStr == "" {
		return -1, nil
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("invalid ID format: %v", err)
	}

	if id < 0 {
		return -1, fmt.Errorf("negative IDs are not allowed")
	}

	return id, nil
}

type Pagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int    `json:"total"`
	Pages    int    `json:"pages"`
	Next     string `json:"next,omitempty"`
}

func (api *API) paginate(total int, page int, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = api.DefaultPageSize
	}
	if pageSize > api.MaxPageSize {
		pageSize = api.MaxPageSize
	}

	pages := (total + pageSize - 1) / pageSize

	pagination := Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Pages:    pages,
	}

	if page < pages {
		pagination.Next = fmt.Sprintf("?page=%d&%s=%d", page+1, PAGESIZE, pageSize)
	}

	return pagination
}
