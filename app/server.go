package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Resp byte

const (
	Integer Resp = ':'
	String  Resp = '+'
	Bulk    Resp = '$'
	Array   Resp = '*'
)

const PORT = 6379
const DELIMITER = "\r\n"

func main() {
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: PORT})
	if err != nil {
		log.Fatalf("Failed to listening on port %v\n", PORT)
	}
	log.Printf("listening on port %v\n", PORT)
	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			continue
		}
		log.Println("Accepted connection from ", conn.RemoteAddr())
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().(*net.TCPAddr).IP.String()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			// in net.Conn, EOF means client connection is closed
			if err == io.EOF {
				log.Printf("%v connection closed\n", addr)
			} else {
				log.Printf("read error at %v connection: %v\n", addr, err.Error())
			}
			return
		}
		log.Printf("Received data from %v: %v\n", addr, string(buf[:n]))
		resp, err := handleMessage(buf[:n])
		if err != nil {
			// TODO handle message read error
			log.Println("TODO")
		}

		fmt.Printf("resp: %v\n", string(resp))
		_, err = conn.Write(resp)
		if err != nil {
			log.Println("write error: ", err.Error())
		}
	}
}

func handleMessage(msg []byte) ([]byte, error) {
	splittedMsg := bytes.Split(msg, []byte(DELIMITER))

	respType := Resp(splittedMsg[0][0])
	switch respType {
	case Array:
		switch strings.ToLower(string(splittedMsg[2])) {
		case "echo":
			result := []byte("+")
			for i := 4; i < len(splittedMsg); i += 2 {
				result = append(result, splittedMsg[i]...)
			}
			result = append(result, []byte(DELIMITER)...)
			return result, nil
		case "ping":
			return []byte("+PONG\r\n"), nil
		}
	}

	return []byte("+hi marceline" + DELIMITER), nil
}
