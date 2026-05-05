package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	req, err := parseReq(reader)

	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println(req.Method, req.Path)

	parseBody(reader, req)

	sendRes(conn, req)
}

type HTTPReq struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Body     []byte
}

func parseReq(reader *bufio.Reader) (*HTTPReq, error) {
	req := HTTPReq{Headers: make(map[string]string)}

	n := 0

	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			return nil, err
		}

		line = bytes.TrimSuffix(line, []byte("\r\n"))

		if len(line) == 0 {
			break
		}

		if n == 0 {
			parts := strings.Fields(string(line))

			if len(parts) != 3 {
				return nil, errors.New("Invalid request")
			}

			req.Method = parts[0]
			req.Path = parts[1]
			req.Protocol = parts[2]
		} else {
			b := bytes.SplitN(line, []byte(": "), 2)
			if len(b) != 2 {
				return nil, errors.New("Corrupt headers")
			}

			key := strings.ToLower(string(bytes.TrimSpace(b[0])))
			val := string(b[1])

			req.Headers[key] = val

		}

		n++
	}

	return &req, nil
}

func parseBody(reader *bufio.Reader, req *HTTPReq) error {
	cl, ok := req.Headers["content-length"]
	if !ok {
		return nil
	}

	length, err := strconv.Atoi(strings.TrimSpace(cl))

	if err != nil {
		fmt.Println("Invalid content length")
	}

	if length > 1<<24 {
		return errors.New("Too long for me")
	}

	body := make([]byte, length)
	_, err = io.ReadFull(reader, body)

	if err != nil {
		return err
	}

	req.Body = body
	return nil
}

func sendRes(conn net.Conn, req *HTTPReq) {
	writer := bufio.NewWriter(conn)

	writer.WriteString("HTTP/1.1 200 OK\r\n")
	for k, v := range req.Headers {
		fmt.Fprintf(writer, "%s: %s\r\n", k, v)
	}
	writer.WriteString("\r\n")

	writer.Write(req.Body)
	writer.Flush()
}
