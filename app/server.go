package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Body    string
	Headers map[string]string
}

const (
	HTTPStatusOK         = "HTTP/1.1 200 OK\r\n\r\n"
	HTTPStatusCreated    = "HTTP/1.1 201 Created\r\n\r\n"
	HTTPStatusNotFound   = "HTTP/1.1 404 Not Found\r\n\r\n"
	BufferSize           = 8192
	DefaultListenAddress = "0.0.0.0:4221"
)

func parseRequest(request string) Request {
	lines := strings.Split(request, "\r\n")

	// Parse the start line
	startLine := lines[0]
	components := strings.Split(startLine, " ")
	method := components[0]
	path := components[1]

	// Parse headers
	headers := make(map[string]string)
	var bodyStartIndex int
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			// Empty line indicates the end of headers, and the next line is the body
			bodyStartIndex = i + 1
			break
		}
		headerComponents := strings.SplitN(line, ": ", 2)
		if len(headerComponents) == 2 {
			// Handle multiline headers
			headers[headerComponents[0]] += headerComponents[1]
		}
	}

	// Parse the body
	body := strings.Join(lines[bodyStartIndex:], "\r\n")

	return Request{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}

func handleConnection(conn net.Conn, directory string) {
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
	if r.Path == "/" {
		response = HTTPStatusOK
	} else if strings.Contains(r.Path, "/echo/") {
		response = HTTPStatusOK
		index := strings.Index(r.Path, "echo/")
		content := r.Path[index+len("echo/"):]
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
	} else if strings.Contains(r.Path, "/user-agent") {
		response = HTTPStatusOK
		content, _ := extractUserAgent(request)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
	} else if r.Method == http.MethodGet && strings.Contains(r.Path, "/files") {
		index := strings.Index(r.Path, "files/")
		fileName := r.Path[index+len("files/"):]

		data, err := readFileIfExists(directory, fileName)
		if err != nil {
			response = HTTPStatusNotFound
		} else {
			response = HTTPStatusOK
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
		}
	} else {
		response = HTTPStatusNotFound
	}

	if r.Method == http.MethodPost {
		// Handle POST request
		body, err := ioutil.ReadAll(strings.NewReader(r.Body))
		index := strings.Index(r.Path, "files/")
		fileName := r.Path[index+len("files/"):]
		if err != nil {
			// Handle the error
			response = HTTPStatusNotFound
		} else {
			response = HTTPStatusCreated
			file, err := os.Create(directory + fileName)
			if err != nil {
				fmt.Println(err)
			}
			_, err = file.Write(body)
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()
		}
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func extractUserAgent(request string) (string, error) {
	// Split the request string into lines
	scanner := bufio.NewScanner(strings.NewReader(request))
	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line starts with "User-Agent:"
		if strings.HasPrefix(line, "User-Agent:") {
			// Extract the User-Agent value
			return strings.TrimSpace(strings.TrimPrefix(line, "User-Agent:")), nil
		}
	}

	// If User-Agent is not found
	return "", fmt.Errorf("User-Agent not found in the request")
}

func readFileIfExists(directory, filename string) ([]byte, error) {
	// Construct the file path
	filePath := fmt.Sprintf("%s/%s", directory, filename)

	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// File does not exist
		return nil, fmt.Errorf("File %s not found in directory %s", filename, directory)
	} else if err != nil {
		// Other error occurred
		return nil, err
	}

	// Read the file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	var dirFlag = flag.String("directory", ".", "directory to serve files from")
	flag.Parse()

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
		// For each accepted connection, launch a goroutine to handle it.
		go handleConnection(conn, *dirFlag)
	}
}
