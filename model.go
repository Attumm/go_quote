package main

type Quote struct {
	Text   string
	Author string
	Tags   []string
}

type ResponseQuote struct {
	Quote
	ID int
}

func (q Quote) CreateResponseQuote(id int) ResponseQuote {
	return ResponseQuote{
		Quote: q,
		ID:    id,
	}
}

type Quotes []Quote
