package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	fmt.Println("TCP server started at 0.0.0.0:4221")

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 8192)
	length, err := conn.Read(buffer)

	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		os.Exit(1)
		os.Exit(1)
	}

	request := string(buffer[:length])
	fmt.Println("Received request:")
	fmt.Println(request)
	r := parseRequest(request)
	var response string
	if r.Path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
	}

	defer conn.Close()
}

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
}

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
