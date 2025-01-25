package main

import (
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []Token
	}{
		{
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
