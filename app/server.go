package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		return
	}

	data := string(buf[:n])

	switch {
	case strings.Contains(data, "PING"):
		conn.Write([]byte("+PONG\r\n"))
	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn)
	}

}
