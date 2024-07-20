package main

import (	
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	fileLocation := getDirLocation()
	fmt.Println("fileLocation: ", fileLocation)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()	
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		
		fmt.Println("going to start a goroutine")
		go handleConnection(conn, fileLocation)
		fmt.Println("after starting a goroutine")	
	}

}

func handleConnection(conn net.Conn, fileLocation string) {
	fmt.Println("starting handle connection function")
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
	
	requestMethod := strings.Split(lines[0], " ")[0]
	fmt.Println("requestMethod: ", requestMethod)
	requestTarget := strings.Split(lines[0], " ")[1]	
	fmt.Println("requestTarget: ", requestTarget)
	

	headersStr := lines[1:len(lines) - 2]   

	requestBody := lines[len(lines) - 1]

	_ = lines[len(lines) - 1]


	fmt.Println("headerStr: ", headersStr)
	var requestHeaders map[string]string = make(map[string]string)
	for _, element := range headersStr {
		colonIndex := strings.Index(element, ":")

		key := strings.TrimSpace(element[:colonIndex])		
		value := strings.TrimSpace(element[colonIndex + 1:])
		requestHeaders[key] = value
		fmt.Println("requestHeaders: key: ", key, ", value: ", value)
	}

	encodingStr := requestHeaders["Accept-Encoding"]
	encodingStr = strings.Replace(encodingStr, " ", "", -1)
	encodings := strings.Split(encodingStr, ",")
	 

	notFoundStatusLine := "HTTP/1.1 404 Not Found"
	okStatusLine := "HTTP/1.1 200 OK"
	createdStatusLine := "HTTP/1.1 201 Created"
	methodNotAllowedStatusLine := "HTTP/1.1 405 Method Not Allowed"


	var statusLine string
	fmt.Println("request target: ", requestTarget)
	var contentTypeStr string = ""
	var contentLengthStr string = ""
	var body string = ""
	var responseHeaders = make(map[string]string)
	if requestTarget == "/" {
		statusLine = okStatusLine	
	} else if strings.HasPrefix(requestTarget, "/echo") {
		statusLine = okStatusLine
		path := strings.Split(requestTarget, "/")[2]
		fmt.Println("path : ", path, len(path))
		body = path
		responseHeaders["Content-Type"] = "text/plain"		
	} else if strings.HasPrefix(requestTarget, "/user-agent") {
		statusLine = okStatusLine

		body = requestHeaders["User-Agent"]
		responseHeaders["Content-Type"] = "text/plain"				
	} else if strings.HasPrefix(requestTarget, "/files") {
		if requestMethod == "GET" {
			slashIndex := strings.Index(requestTarget, "/files/")
			filename := requestTarget[slashIndex + len("/files/"):]
			fmt.Println("filename: ", filename)
	
			dat, err := os.ReadFile(fileLocation + filename)
			if err != nil {
				fmt.Println("unable to read ", filename, ". err: ", err.Error())
	
				statusLine = notFoundStatusLine
			} else {
				fmt.Println("file content: ", string(dat))
				statusLine = okStatusLine
				body = string(dat)
				responseHeaders["Content-Type"] = "application/octet-stream"		
			}	
		} else if requestMethod == "POST" {
			slashIndex := strings.Index(requestTarget, "/files/")
			filename := requestTarget[slashIndex + len("/files/"):]

			os.WriteFile(fileLocation + filename, []byte(requestBody), fs.ModeAppend)
			
			statusLine = createdStatusLine
		} else {
			fmt.Println("not supported method: ", requestMethod)
			statusLine = methodNotAllowedStatusLine
		}
	} else {
		statusLine = notFoundStatusLine
	}
	fmt.Println("statusLine: ", statusLine, ", contentType: ", contentTypeStr, ", contentLength: ", contentLengthStr, ", body: ", body)
	if slices.Contains(encodings, "gzip") {
		responseHeaders["Content-Encoding"] = "gzip"
	}
	var bodyBytes []byte
	if body != "" {	
		if slices.Contains(encodings, "gzip") {
			zipContent, err := GzipData([]byte(body))
			if err != nil {
				fmt.Println("unable to gzip data: ", err.Error())
				return
			}	
			fmt.Println("zip content: ", zipContent.String())
			bodyBytes = zipContent.Bytes()			
		} else {
			bodyBytes = []byte(body)
		}
	}

	response := statusLine + "\r\n"
	if bodyBytes != nil {		
		responseHeaders["Content-Length"] = strconv.Itoa(len(bodyBytes))
	}
	
	// response += "\r\n"
	// if content != "" {	
	// 	if slices.Contains(encodings, "gzip") {
	// 		zipContent, err := GzipData(content)
	// 		fmt.Println("zip: ", zipContent)
	// 		if err != nil {
	// 			fmt.Println("unable to gzip data: ", err.Error())
	// 			return
	// 		}	
	// 		fmt.Println("zip content: ", zipContent.String())
	// 		response += zipContent.String()
	// 	} else {
	// 		response += content
	// 	}
	// }

	var responseByteArray = make([]byte, 0)
	responseByteArray = append(responseByteArray, []byte(statusLine)...)
	responseByteArray = append(responseByteArray, []byte("\r\n")...)
	for key, value := range responseHeaders {
		header := key + ": " + value
		fmt.Println("responseHeader: ", header)
		responseByteArray = append(responseByteArray, []byte(header)...)
		responseByteArray = append(responseByteArray, []byte("\r\n")...)
	}
	responseByteArray = append(responseByteArray, []byte("\r\n")...)
	if (bodyBytes != nil) {
		responseByteArray = append(responseByteArray, bodyBytes...)
	}
	
	response2 := "\"" + response + "\""
	fmt.Println("Writing response to conn: ")
	fmt.Println(response2)
	// responseByteArray := []byte(response)
	// var responseByteArray []byte
	fmt.Println(string(responseByteArray))
	n, err  := conn.Write(responseByteArray)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("number written: ", n)
}

func GzipData(data []byte) (bytes.Buffer, error) {
	fmt.Println("gziping ", data, string(data))
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)

	if err != nil {
		return buf, err
	}

	if err := zw.Close(); err != nil {
		fmt.Println("cannot close zw: ", err.Error())
		return buf, err	
	}
	
	return buf, err
}

func getDirLocation() string {
	args := os.Args[1:]
	var directoryLocation string
	if len(args) == 0 {
		fmt.Println("You need to specify directory location by --directory")
		// os.Exit(1)
		return ""
	}
	if args[0] == "--directory" {
		if len(args) == 1 {
			fmt.Println("You need to specify directory location value after --directory")
			// os.Exit(1)
			return ""
		}
		directoryLocation = args[1]
	}

	return directoryLocation
}