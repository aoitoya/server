package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
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

	req, err := parseReq(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println(req.Method, req.Path)

	sendRes(conn, req)
}

type HTTPReq struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
}

func parseReq(conn net.Conn) (*HTTPReq, error) {
	reader := bufio.NewReader(conn)

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
			b := bytes.Split(line, []byte{' '})

			if len(b) != 3 {
				return nil, errors.New("Invalid request")
			}

			req.Method = string(b[0])
			req.Path = string(b[1])
			req.Protocol = string(b[2])
		} else {
			b := bytes.Split(line, []byte(": "))
			if len(b) != 2 {
				return nil, errors.New("Corrupt headers")
			}

			key := string(b[0])
			val := string(b[1])

			req.Headers[key] = val

		}

		n++
	}

	return &req, nil
}

func sendRes(conn net.Conn, req *HTTPReq) {
	writer := bufio.NewWriter(conn)

	writer.WriteString("HTTP/1.1 200 OK\r\n")
	writer.WriteString("\r\n")
	for k, v := range req.Headers {
		fmt.Fprintf(writer, "%s: %s\r\n", k, v)
	}
	writer.Flush()
}
