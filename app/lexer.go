package main

import (
	"fmt"
	"strconv"
)

const (
	TokenArray      = "*"
	TokenBulkString = "$"
	TokenSimpleStr  = "+"
	TokenError      = "-"
	TokenInteger    = ":"
	TokenString     = "string"
	TokenCRLF       = "\r\n"
)

type Token struct {
	Type  string
	Value string
}

type Lexer struct {
	input        string
	current      int
	state        string
	length       int
	currentToken Token
	currentValue string
	tokens       []Token
}

func (l *Lexer) ReadChar() rune {
	char := rune(l.input[l.current])
	l.current++

	return char
}

func (l *Lexer) CreateToken(tokenType string) {
	if tokenType == TokenBulkString {
		l.state = "reading_length"
	}

	if tokenType == TokenCRLF {
		l.currentValue = "\r\n"
	}

	l.currentToken = Token{Type: tokenType}
}

func (l *Lexer) PushCurrentToken() {
	l.currentToken.Value = l.currentValue
	l.currentValue = ""

	l.tokens = append(l.tokens, l.currentToken)
}

func (l *Lexer) ReadBulkString() {
	l.currentValue = l.input[l.current-1 : l.current+l.length-1]
	l.current += (l.length - 1)

	l.state = "initial"
}

func (l *Lexer) Tokenize() []Token {
	for l.current < len(l.input) {
		char := l.ReadChar()

		switch char {
		case '*':
			l.CreateToken(TokenArray)
		case '$':
			l.CreateToken(TokenBulkString)
		case '\r':
			l.PushCurrentToken()

			if l.ReadChar() == '\n' {
				l.CreateToken(TokenCRLF)
				l.PushCurrentToken()

				if l.state == "reading_length" {
					l.state = "reading_string"
				}

			} else {
				panic(fmt.Sprintf("Unexpected character after \r: %v", char))
			}
		default:
			switch l.state {
			case "reading_string":
				l.CreateToken(TokenString)
				l.ReadBulkString()
			case "reading_length":
				val, err := strconv.Atoi(string(char))
				if err != nil {
					panic(fmt.Sprintf("Error on converting string to int: %v", char))
				}

				l.length = val
				l.currentValue += string(char)
			default:
				l.currentValue += string(char)
			}
		}
	}

	return l.tokens
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, current: 0, state: "initial"}
}
