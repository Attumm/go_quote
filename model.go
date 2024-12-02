package main

import "net/url"

type Quote struct {
	Text   string
	Author string
	Tags   []string
}

type ResponseQuote struct {
	Quote
	ID       int
	AuthorID string
}

func (q Quote) CreateResponseQuote(id int) ResponseQuote {
	return ResponseQuote{
		Quote:    q,
		ID:       id,
		AuthorID: url.QueryEscape(q.Author),
	}
}

type Quotes []Quote

type IndexStructure struct {
	names     []string
	nameToIDs map[string][]int
}

func NewIndexStructure() IndexStructure {
	return IndexStructure{
		names:     make([]string, 0),
		nameToIDs: make(map[string][]int),
	}
}

func (is *IndexStructure) add(name string, id int) {
	if _, exists := is.nameToIDs[name]; !exists {
		is.names = append(is.names, name)
	}
	is.nameToIDs[name] = append(is.nameToIDs[name], id)
}

func BuildAuthorIndex(quotes Quotes) IndexStructure {
	index := NewIndexStructure()
	for i, quote := range quotes {
		index.add(url.QueryEscape(quote.Author), i)
	}
	return index
}

func BuildTagIndex(quotes Quotes) IndexStructure {
	index := NewIndexStructure()
	for i, quote := range quotes {
		for _, tag := range quote.Tags {
			index.add(tag, i)
		}
	}
	return index
}
