package main

import (
	// "bufio"
	// "bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()	
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	readCount, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("unable to read from connection: ", err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("read count: ", readCount)

	readData := string(buf[:readCount])
	fmt.Println("readData: ", readData)

	lines := strings.Split(readData, "\r\n")
	fmt.Println("lines: ", lines)
	
	requestTarget := strings.Split(lines[0], " ")[1]	

	notFoundStatusLine := "HTTP/1.1 404 Not Found"
	okStatusLine := "HTTP/1.1 200 OK"

	var statusLine string
	fmt.Println("request target: ", requestTarget)
	var contentTypeStr string = ""
	var contentLengthStr string = ""
	var content string = ""
	if requestTarget == "/" {
		statusLine = okStatusLine	
	} else if strings.HasPrefix(requestTarget, "/echo") {
		statusLine = okStatusLine
		path := strings.Split(requestTarget, "/")[2]
		fmt.Println("path : ", path, len(path))
		content = path
		contentTypeStr = "Content-Type: text/plain"
		contentLength := strconv.Itoa(len(content))		
		contentLengthStr = "Content-Length: " + contentLength
	} else {
		statusLine = notFoundStatusLine
	}
	fmt.Println("statusLine: ", statusLine, ", contentType: ", contentTypeStr, ", contentLength: ", contentLengthStr, ", content: ", content)
	// response := statusLine + "\r\n\r\n" + contentTypeStr + "\r\n" + contentLengthStr + "\r\n\r\n" + content
	response := statusLine + "\r\n" + contentTypeStr + "\r\n" + contentLengthStr + "\r\n\r\n" + content
	response2 := "\"" + response + "\""
	fmt.Println("Writing response to conn: ")
	fmt.Println(response2)
	responseByteArray := []byte(response)
	n, err  := conn.Write(responseByteArray)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("number written: ", n)
}
