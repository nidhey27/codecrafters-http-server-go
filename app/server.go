package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
}

const (
	HTTPStatusOK         = "HTTP/1.1 200 OK\r\n\r\n"
	HTTPStatusNotFound   = "HTTP/1.1 404 Not Found\r\n\r\n"
	BufferSize           = 8192
	DefaultListenAddress = "0.0.0.0:4221"
)

func parseRequest(request string) Request {
	lines := strings.Split(request, "\r\n")
	startLine := lines[0]
	components := strings.Split(startLine, " ")
	method := components[0]
	path := components[1]
	return Request{
		Method:  method,
		Path:    path,
		Headers: make(map[string]string),
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, BufferSize)
	length, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading:", err)
		return
	}

	request := string(buffer[:length])
	fmt.Println("Received request:")
	fmt.Println(request)

	r := parseRequest(request)

	var response string
	responseBody := "ABC"
	if r.Path == "/" {
		response = HTTPStatusOK
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(responseBody), responseBody)
	} else {
		response = HTTPStatusNotFound
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func main() {
	l, err := net.Listen("tcp", DefaultListenAddress)
	if err != nil {
		log.Fatal("Failed to bind to port 4221:", err)
	}

	defer l.Close()

	fmt.Println("TCP server started at", DefaultListenAddress)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
