package main

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "array with bulk strings",
			input: "*2\r\n$4\r\nECHO\r\n$3\r\nHEY\r\n",
			expected: []Token{
				{Type: TokenArray, Value: "2"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "4"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "ECHO"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "3"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "HEY"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
		},
		{
			name:  "simple string",
			input: "+OK\r\n",
			expected: []Token{
				{Type: TokenSimpleStr, Value: "OK"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
		},
		{
			name:  "error message",
			input: "-Error occurred\r\n",
			expected: []Token{
				{Type: TokenError, Value: "Error occurred"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
		},
		{
			name:  "integer",
			input: ":1000\r\n",
			expected: []Token{
				{Type: TokenInteger, Value: "1000"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
		},
		{
			name:  "complex command",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			expected: []Token{
				{Type: TokenArray, Value: "3"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "3"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "SET"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "3"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "key"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "5"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "value"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
		},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(tokens) != len(tt.expected) {
			t.Errorf("wrong number of tokens, got=%d, want=%d",
				len(tokens), len(tt.expected))
			continue
		}

		for i, tok := range tokens {
			if tok != tt.expected[i] {
				t.Errorf("wrong token at position %d\ngot=%+v\nwant=%+v",
					i, tok, tt.expected[i])
			}
		}
	}
}
