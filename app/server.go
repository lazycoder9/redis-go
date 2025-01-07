package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var storage = make(map[string]Record)

var commands map[string]func(r *Request) = map[string]func(r *Request){
	"ECHO": handleEchoCommand,
	"PING": handlePingCommand,
	"SET":  handleSetCommand,
	"GET":  handleGetCommand,
}

func handleEchoCommand(r *Request) {
	parts := Chunk(r.Fields[1:], 2)
	r.Conn.Write([]byte(strings.Join(parts[1], "\r\n") + "\r\n"))
}

func handleUnknownError(r *Request) {
	r.Conn.Write([]byte("-ERR unknown command\r\n"))
}

func handlePingCommand(r *Request) {
	r.Conn.Write([]byte("+PONG\r\n"))
}

func handleSetCommand(r *Request) {
	parts := Chunk(r.Fields[1:], 2)
	key := parts[1][1]
	value := parts[2]
	options := parts[3:]
	record := Record{value: value}

	for i := 0; i < len(options); i += 2 {
		optionName := strings.ToUpper(options[i][1])
		optionValue := options[i+1][1]

		if optionName == "PX" {
			ms, err := strconv.ParseInt(optionValue, 10, 64)
			if err != nil {
				r.Conn.Write([]byte("-ERR cannot parse PX value"))
			}

			duration := time.Duration(ms) * time.Millisecond
			record.expiration = time.Now().Add(duration)
		}

	}
	storage[key] = record
	r.Conn.Write([]byte("+OK\r\n"))
}

func handleGetCommand(r *Request) {
	parts := Chunk(r.Fields[1:], 2)
	key := parts[1][1]
	record, recordExist := storage[key]

	if !recordExist {
		r.Conn.Write([]byte("$-1\r\n"))
		return
	}

	if !record.expiration.IsZero() && time.Now().After(record.expiration) {
		delete(storage, key)
		r.Conn.Write([]byte("$-1\r\n"))
		return
	}

	v := record.value.([]string)
	r.Conn.Write([]byte(strings.Join(v, "\r\n") + "\r\n"))
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
	Type   string
	Raw    []byte
	Conn   net.Conn
	Fields []string
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

func (r *Request) parseType() {
	var requestType string
	switch string(r.Raw[0]) {
	case "$":
		requestType = "bulk"
	case "+":
		requestType = "string"
	case "*":
		requestType = "array"
	default:
		requestType = ""
	}

	r.Type = requestType
}

func (r *Request) handleStringRequest() {
	switch {
	case strings.HasPrefix(string(r.Raw[1:]), "PING"):
		r.Conn.Write([]byte("+PONG\r\n"))
	default:
		r.Conn.Write([]byte("-ERR unknown command\r\n"))
	}
}

func (r *Request) handleArrayRequest() {
	parts := Chunk(r.Fields[1:], 2)
	command := strings.ToUpper(parts[0][1])

	if handler, ok := commands[command]; ok {
		handler(r)
	} else {
		handleUnknownError(r)
	}
}

func makeRequest(conn net.Conn, rawData []byte) Request {
	fields := strings.Fields(string(rawData))
	req := Request{Conn: conn, Raw: rawData, Fields: fields}
	return req
}

func handleRequest(r Request) {
	switch r.Type {
	case "string":
		r.handleStringRequest()
	case "array":
		r.handleArrayRequest()
	default:
		r.Conn.Write([]byte("-ERR unknown command\r\n"))
	}
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
		request := makeRequest(conn, data)
		request.parseType()
		handleRequest(request)
	}
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}
