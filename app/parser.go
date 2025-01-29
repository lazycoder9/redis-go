package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Command struct {
	Name string
	Args []string
}

type (
	ParserStateChecker func(p *Parser) error
	ParserProcessor    func(p *Parser, c *Command) error
)

const (
	parserStateInitial               = "initial"                     // Initial state. No tokens are parsed yet. We expect array token as first element
	parserStateReadCommandLength     = "reading_command_length"      // Read command
	parserStateReadCommand           = "reading_command"             // Read command
	parserStateReadCommandArgsLength = "reading_command_args_length" // Read command args
	parserStateReadCommandArgs       = "reading_command_args"        // Read command args
)

var (
	validCommands = []string{"ECHO", "PING", "SET"}
	ignoredTokens = []string{TokenCRLF, TokenBulkString}
)

var checkers map[string]ParserStateChecker = map[string]ParserStateChecker{
	parserStateInitial: func(p *Parser) error {
		if p.tokens[p.pos].Type != TokenArray {
			return fmt.Errorf("Invalid: no array start")
		}
		return nil
	},
	parserStateReadCommandLength: func(p *Parser) error {
		if p.tokens[p.pos].Type != TokenBulkString {
			return fmt.Errorf("Invalid: expected command string length")
		}

		_, err := strconv.Atoi(p.tokens[p.pos].Value)
		if err != nil {
			return fmt.Errorf("Invalid: wrong command length attribute %v", p.tokens[p.pos].Value)
		}

		return nil
	},
	parserStateReadCommand: func(p *Parser) error {
		if p.tokens[p.pos].Type != TokenString {
			return fmt.Errorf("Invalid: expected string as command")
		}

		return nil
	},
	parserStateReadCommandArgsLength: func(p *Parser) error {
		if p.tokens[p.pos].Type != TokenBulkString {
			return fmt.Errorf("Invalid: expected command arg string length")
		}

		_, err := strconv.Atoi(p.tokens[p.pos].Value)
		if err != nil {
			return fmt.Errorf("Invalid: wrong command arg length attribute %v", p.tokens[p.pos].Value)
		}

		return nil
	},
	parserStateReadCommandArgs: func(p *Parser) error {
		if p.tokens[p.pos].Type != TokenString {
			return fmt.Errorf("invalid: expected string as command arg")
		}

		return nil
	},
}

var processors map[string]ParserProcessor = map[string]ParserProcessor{
	parserStateInitial: func(p *Parser, c *Command) error {
		arrayLength, err := strconv.Atoi(p.tokens[p.pos].Value)
		if err != nil {
			return fmt.Errorf("invalid: wrong array length attribute %v", p.tokens[p.pos].Value)
		}

		p.expectedLength = arrayLength
		p.pos++
		p.state = stateTransitions[p.state]
		return nil
	},
	parserStateReadCommandLength: func(p *Parser, c *Command) error {
		p.pos++
		p.state = stateTransitions[p.state]
		return nil
	},
	parserStateReadCommand: func(p *Parser, c *Command) error {
		command := p.tokens[p.pos].Value

		if !slices.Contains(validCommands, command) {
			return fmt.Errorf("invalid command: %v", command)
		}

		c.Name = strings.ToUpper(command)
		p.pos++
		p.state = stateTransitions[p.state]
		p.actualLength++
		return nil
	},
	parserStateReadCommandArgsLength: func(p *Parser, c *Command) error {
		p.pos++
		p.state = stateTransitions[p.state]
		return nil
	},
	parserStateReadCommandArgs: func(p *Parser, c *Command) error {
		c.Args = append(c.Args, p.tokens[p.pos].Value)
		p.pos++
		p.state = stateTransitions[p.state]
		p.actualLength++
		return nil
	},
}

var stateTransitions map[string]string = map[string]string{
	parserStateInitial:               parserStateReadCommandLength,
	parserStateReadCommandLength:     parserStateReadCommand,
	parserStateReadCommand:           parserStateReadCommandArgsLength,
	parserStateReadCommandArgsLength: parserStateReadCommandArgs,
	parserStateReadCommandArgs:       parserStateReadCommandArgsLength,
}

type Parser struct {
	tokens         []Token
	pos            int // Current position in tokens
	state          string
	expectedLength int
	actualLength   int
}

func (p *Parser) IsFinished() bool {
	return p.pos >= len(p.tokens)
}

func (p *Parser) Parse() (*Command, error) {
	command := &Command{}
	for !p.IsFinished() {

		if err := checkers[p.state](p); err != nil {
			return nil, err
		}

		if err := processors[p.state](p, command); err != nil {
			return nil, err
		}

		if p.tokens[p.pos].Type == TokenCRLF {
			p.pos++
		}
	}

	if p.expectedLength != p.actualLength {
		return nil, fmt.Errorf("wrong array length, expected: %d, actual: %d", p.expectedLength, p.actualLength)
	}

	return command, nil
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0, state: parserStateInitial}
}
