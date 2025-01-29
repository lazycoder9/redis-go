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
	stateArrayHeader   = "array_header"   // First token (*N) indicating number of array elements
	stateCommandLength = "command_length" // Length of command name ($N)
	stateCommandName   = "command_name"   // Actual command name (GET, SET, etc.)
	stateArgLength     = "arg_length"     // Length of command argument ($N)
	stateArgValue      = "arg_value"      // Actual argument value
)

type ParserState struct {
	validate  ParserStateChecker
	process   ParserProcessor
	nextState string
}

var validCommands = []string{"ECHO", "PING", "SET"}

type Parser struct {
	tokens         []Token
	pos            int
	state          string
	expectedLength int
	actualLength   int
}

var parserStates = map[string]ParserState{
	stateArrayHeader: {
		validate: func(p *Parser) error {
			if p.tokens[p.pos].Type != TokenArray {
				return fmt.Errorf("expected array token, got %s", p.tokens[p.pos].Type)
			}
			return nil
		},
		process: func(p *Parser, c *Command) error {
			arrayLength, err := strconv.Atoi(p.tokens[p.pos].Value)
			if err != nil {
				return fmt.Errorf("invalid array length: %s", p.tokens[p.pos].Value)
			}
			p.expectedLength = arrayLength
			return nil
		},
		nextState: stateCommandLength,
	},
	stateCommandLength: {
		validate: func(p *Parser) error {
			if p.tokens[p.pos].Type != TokenBulkString {
				return fmt.Errorf("expected bulk string token for command length")
			}
			_, err := strconv.Atoi(p.tokens[p.pos].Value)
			if err != nil {
				return fmt.Errorf("invalid command length: %s", p.tokens[p.pos].Value)
			}
			return nil
		},
		process: func(p *Parser, c *Command) error {
			return nil // Just validate length, actual command comes next
		},
		nextState: stateCommandName,
	},
	stateCommandName: {
		validate: func(p *Parser) error {
			if p.tokens[p.pos].Type != TokenString {
				return fmt.Errorf("expected string token for command name")
			}
			if !slices.Contains(validCommands, strings.ToUpper(p.tokens[p.pos].Value)) {
				return fmt.Errorf("invalid command: %s", p.tokens[p.pos].Value)
			}
			return nil
		},
		process: func(p *Parser, c *Command) error {
			c.Name = strings.ToUpper(p.tokens[p.pos].Value)
			p.actualLength++
			return nil
		},
		nextState: stateArgLength,
	},
	stateArgLength: {
		validate: func(p *Parser) error {
			if p.tokens[p.pos].Type != TokenBulkString {
				return fmt.Errorf("expected bulk string token for argument length")
			}
			_, err := strconv.Atoi(p.tokens[p.pos].Value)
			if err != nil {
				return fmt.Errorf("invalid argument length: %s", p.tokens[p.pos].Value)
			}
			return nil
		},
		process: func(p *Parser, c *Command) error {
			return nil // Just validate length, actual argument comes next
		},
		nextState: stateArgValue,
	},
	stateArgValue: {
		validate: func(p *Parser) error {
			if p.tokens[p.pos].Type != TokenString {
				return fmt.Errorf("expected string token for argument value")
			}
			return nil
		},
		process: func(p *Parser, c *Command) error {
			c.Args = append(c.Args, p.tokens[p.pos].Value)
			p.actualLength++
			return nil
		},
		nextState: stateArgLength,
	},
}

func (p *Parser) IsFinished() bool {
	return p.pos >= len(p.tokens)
}

func (p *Parser) advanceToken() {
	p.pos++

	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenCRLF {
		p.pos++
	}
}

func (p *Parser) Parse() (*Command, error) {
	command := &Command{}

	for !p.IsFinished() {
		state, exists := parserStates[p.state]
		if !exists {
			return nil, fmt.Errorf("invalid parser state: %s", p.state)
		}

		if err := state.validate(p); err != nil {
			return nil, err
		}

		if err := state.process(p, command); err != nil {
			return nil, err
		}

		p.advanceToken()
		p.state = state.nextState
	}

	if p.expectedLength != p.actualLength {
		return nil, fmt.Errorf("wrong array length, expected: %d, actual: %d",
			p.expectedLength, p.actualLength)
	}

	return command, nil
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		state:  stateArrayHeader,
	}
}
