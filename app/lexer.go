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

	stateInitial       = "initial"
	stateReadingLength = "reading_length"
	stateReadingString = "reading_string"
)

type Token struct {
	Type  string
	Value string
}

type tokenBuffer struct {
	token Token
	value string
}

type Lexer struct {
	input       string
	current     int
	state       string
	length      int
	tokenBuffer tokenBuffer
	tokens      []Token
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
		l.tokenBuffer.value = "\r\n"
	}

	l.tokenBuffer.token = Token{Type: tokenType}
}

func (l *Lexer) PushCurrentToken() {
	l.tokenBuffer.token.Value = l.tokenBuffer.value
	l.tokenBuffer.value = ""

	l.tokens = append(l.tokens, l.tokenBuffer.token)
}

func (l *Lexer) ReadBulkString() error {
	endPos := l.current + l.length - 1
	if endPos > len(l.input) {
		return fmt.Errorf("bulk string length %d exceeds input length at position %d", l.length, l.current)
	}

	l.tokenBuffer.value = l.input[l.current-1 : endPos]
	l.current = endPos
	l.state = "initial"
	return nil
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

				if l.state == stateReadingLength {
					l.state = stateReadingString
				}

			} else {
				panic(fmt.Sprintf("Unexpected character after \r: %v", char))
			}
		default:
			switch l.state {
			case stateReadingString:
				l.CreateToken(TokenString)
				if err := l.ReadBulkString(); err != nil {
					panic(fmt.Sprintf("Error on reading bulk string with length %d", l.length))
				}
			case stateReadingLength:
				val, err := strconv.Atoi(string(char))
				if err != nil {
					panic(fmt.Sprintf("Error on converting string to int: %v", char))
				}

				l.length = val
				l.tokenBuffer.value += string(char)
			default:
				l.tokenBuffer.value += string(char)
			}
		}
	}

	return l.tokens
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, current: 0, state: stateInitial}
}
