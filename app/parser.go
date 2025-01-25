package main

type Command struct {
	Name string
	Args []string
}

type Parser struct {
	tokens []Token
	pos    int // Current position in tokens
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}
