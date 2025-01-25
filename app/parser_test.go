package main

import "testing"

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []Token
		expected *Command
		wantErr  bool
	}{
		{
			name: "ECHO command",
			tokens: []Token{
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
			expected: &Command{Name: "ECHO", Args: []string{"HEY"}},
		},
		{
			name: "SET command",
			tokens: []Token{
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
			expected: &Command{Name: "SET", Args: []string{"key", "value"}},
		},
		{
			name: "PING command no args",
			tokens: []Token{
				{Type: TokenArray, Value: "1"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "4"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "PING"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
			expected: &Command{Name: "PING", Args: []string{}},
		},
		{
			name: "Invalid: no array start",
			tokens: []Token{
				{Type: TokenString, Value: "PING"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
			wantErr: true,
		},
		{
			name: "Invalid: wrong array length",
			tokens: []Token{
				{Type: TokenArray, Value: "3"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenBulkString, Value: "4"},
				{Type: TokenCRLF, Value: "\r\n"},
				{Type: TokenString, Value: "PING"},
				{Type: TokenCRLF, Value: "\r\n"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.tokens)
			cmd, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cmd.Name != tt.expected.Name {
				t.Errorf("wrong command name, got=%s, want=%s", cmd.Name, tt.expected.Name)
			}

			if len(cmd.Args) != len(tt.expected.Args) {
				t.Fatalf("wrong number of args, got=%d, want=%d", len(cmd.Args), len(tt.expected.Args))
			}

			for i, arg := range cmd.Args {
				if arg != tt.expected.Args[i] {
					t.Errorf("wrong arg at position %d, got=%s, want=%s", i, arg, tt.expected.Args[i])
				}
			}
		})
	}
}
