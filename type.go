package main

type Heading struct {
	Tag   string
	Level int
	Count int
}
type Links struct {
	Inaccessible int
	External     int
	Internal     int
}

type DOMAnalysis struct {
	Title     string
	Headings  []Heading
	Links     Links
	LoginForm bool
}

type PageAnalysis struct {
	DOMAnalysis
	HTMLVersion string
}
