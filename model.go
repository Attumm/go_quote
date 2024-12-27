package main

import (
	"net/url"
	"strings"
)

type Category int

const (
	QuotesTypeRequest Category = iota
	AuthorsTypeRequest
	TagsTypeRequest
)

type Quote struct {
	Text   string   `json:"text"`
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

type ResponseQuote struct {
	Quote
	ID       int    `json:"id"`
	AuthorID string `json:"author_id"`
}

func (q Quote) CreateResponseQuote(id int) ResponseQuote {
	authorID := url.QueryEscape(q.Author)
	return ResponseQuote{
		Quote:    q,
		ID:       id,
		AuthorID: authorID,
	}
}

type Quotes []Quote

type ResponseQuotes []ResponseQuote

type IndexStructure struct {
	Names        []string
	NameToQuotes map[string][]int
}

func NewIndexStructure() IndexStructure {
	return IndexStructure{
		Names:        make([]string, 0),
		NameToQuotes: make(map[string][]int),
	}
}

func (is *IndexStructure) Add(name string, id int) {
	parsedName := strings.TrimSpace(name)
	if len(parsedName) == 0 {
		return
	}

	if _, exists := is.NameToQuotes[parsedName]; !exists {
		is.Names = append(is.Names, parsedName)
	}
	is.NameToQuotes[parsedName] = append(is.NameToQuotes[parsedName], id)
}

func (is *IndexStructure) Len() int {
	return len(is.Names)
}

func (cat Category) String() string {
	return [...]string{"QuotesTypeRequest", "AuthorsTypeRequest", "TagsTypeRequest"}[cat]
}

func BuildAuthorIndex(quotes Quotes) IndexStructure {
	index := NewIndexStructure()
	for i, quote := range quotes {
		index.Add(url.QueryEscape(quote.Author), i)
	}
	return index
}

func BuildTagIndex(quotes Quotes) IndexStructure {
	index := NewIndexStructure()
	for i, quote := range quotes {
		for _, tag := range quote.Tags {
			index.Add(tag, i)
		}
	}
	return index
}
