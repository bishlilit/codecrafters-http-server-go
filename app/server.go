package main

import (
	// "bufio"
	// "bytes"
	"fmt"
	"io"
	"net"
	"os"
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
	defer conn.Close()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

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

	notFoundStatusLine := "HTTP/1.1 404 Not Found\r\n\r\n"
	okStatusLine := "HTTP/1.1 200 OK\r\n\r\n"

	var statusLine string
	fmt.Println("request target: ", requestTarget)
	if requestTarget == "/" {
		statusLine = okStatusLine	
	} else {
		statusLine = notFoundStatusLine
	}
	statusLineByteArray := []byte(statusLine)
	fmt.Println("Writing statusLine to conn: ", statusLine)
	n, err  := conn.Write(statusLineByteArray)
	if err != nil {
		fmt.Println("Error writing status line: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("number written: ", n)
}
