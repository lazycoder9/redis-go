package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	storage          = make(map[string]Record)
	configDBFileName string
	configDir        string
)

var commands map[string]CommandHandler = map[string]CommandHandler{
	"ECHO":   handleEchoCommand,
	"PING":   handlePingCommand,
	"SET":    handleSetCommand,
	"GET":    handleGetCommand,
	"CONFIG": handleConfigCommand,
}

type CommandHandler func(c *Command) string

func handleEchoCommand(c *Command) string {
	return c.Args[0]
}

func handleUnknownError(c *Command) string {
	return "-ERR nunknown command"
}

func handlePingCommand(c *Command) string {
	return "PONG"
}

func handleConfigCommand(c *Command) string {
	fmt.Printf("Handle config command: %+v\n", c)

	return ""
}

func handleSetCommand(c *Command) string {
	key := c.Args[0]
	value := c.Args[1]
	// _options := c.Args[2:]

	record := Record{value: value}

	// for i := 0; i < len(options); i += 2 {
	// 	optionName := strings.ToUpper(options[i][1])
	// 	optionValue := options[i+1][1]
	//
	// 	if optionName == "PX" {
	// 		ms, err := strconv.ParseInt(optionValue, 10, 64)
	// 		if err != nil {
	// 			r.Conn.Write([]byte("-ERR cannot parse PX value"))
	// 		}
	//
	// 		duration := time.Duration(ms) * time.Millisecond
	// 		record.expiration = time.Now().Add(duration)
	// 	}
	//
	// }

	storage[key] = record

	return "+OK"
}

func handleGetCommand(c *Command) string {
	key := c.Args[0]
	record, recordExist := storage[key]

	if !recordExist {
		return "$-1"
	}

	// if !record.expiration.IsZero() && time.Now().After(record.expiration) {
	// 	delete(storage, key)
	// 	r.Conn.Write([]byte("$-1\r\n"))
	// 	return
	// }

	v := record.value.([]string)
	return strings.Join(v, "\r\n")
}

type Record struct {
	value      interface{}
	expiration time.Time
}

type Response struct {
	Type string
	Data []byte
}

type Request struct {
	Raw     []byte
	Conn    net.Conn
	Command *Command
}

func (r *Response) buildResponseMessage() string {
	return string(r.Data)
}

func (r *Response) getPrefix() string {
	switch r.Type {
	case "bulk":
		return "$4"
	default:
		return "+"
	}
}

func (r *Request) handleStringRequest() {
	switch {
	case strings.HasPrefix(string(r.Raw[1:]), "PING"):
		r.Conn.Write([]byte("+PONG\r\n"))
	default:
		r.Conn.Write([]byte("-ERR unknown command\r\n"))
	}
}

func makeRequest(conn net.Conn, rawData []byte, command *Command) Request {
	req := Request{Conn: conn, Raw: rawData, Command: command}

	return req
}

func handleCommand(command *Command) string {
	if commandHandler, exist := commands[command.Name]; exist {
		return commandHandler(command)
	}

	return "-ERR unknown command"
}

func handleRequest(r *Request) {
	if r.Command == nil {
		r.Conn.Write([]byte("-ERR unknown command\r\n"))
		return
	}

	response := handleCommand(r.Command)
	fmt.Println("response: ", response)

	r.Conn.Write([]byte("-ERR unknown command\r\n"))
}

func handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Error reading: ", err.Error())
			return
		}
		data := buf[:n]

		lexer := NewLexer(string(data))
		tokens, err := lexer.Tokenize()
		if err != nil {
			fmt.Errorf("-ERR on lexer: %s", err)
		}

		parser := NewParser(tokens)
		cmd, err := parser.Parse()
		if err != nil {
			fmt.Errorf("-ERR on parser: %s", err)
		}

		request := makeRequest(conn, data, cmd)
		handleRequest(&request)
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer listener.Close()

	flag.StringVar(&configDir, "dir", "./", "dir config")
	flag.StringVar(&configDBFileName, "dbfilename", "dump.rdb", "dbFileName config")
	flag.Parse()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
