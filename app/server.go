package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
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

var db map[string]string = map[string]string{}

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
			log.Println(err.Error())
		}

		fmt.Printf("resp: %v\n", string(resp))
		_, err = conn.Write(resp)
		if err != nil {
			log.Println("write error: ", err.Error())
		}
	}
}

func handleMessage(msg []byte) (result []byte, err error) {
	splittedMsg := bytes.Split(msg, []byte(DELIMITER))

	for _, v := range splittedMsg {
		fmt.Printf("v: %v\n", string(v))
	}

	result = []byte{byte(String)}

	respType := Resp(splittedMsg[0][0])
	switch respType {
	case Array:
		arrayLen, _ := strconv.Atoi(string(splittedMsg[0][1:]))
		switch strings.ToLower(string(splittedMsg[2])) {
		case "echo":
			for i := 4; i < len(splittedMsg); i += 2 {
				result = append(result, splittedMsg[i]...)
			}

		case "ping":
			result = append(result, []byte("PONG")...)

		case "set":
			key, value := string(splittedMsg[4]), string(splittedMsg[6])
			db[key] = value

			if arrayLen > 3 && strings.ToLower(string(splittedMsg[8])) == "px" {
				duration, err := strconv.Atoi(string(splittedMsg[10]))
				if err != nil {
					return nil, err
				}
				go handlePx(time.After(time.Millisecond*time.Duration(duration)), key)
			}

			result = append(result, []byte("OK")...)

		case "get":
			value, ok := db[string(splittedMsg[4])]
			if ok {
				result = append(result, []byte(value)...)
			} else {
				result = append(result, []byte("(nil)")...)
				err = fmt.Errorf("key not found")
			}

		default:
			result = append(result, []byte("hi marceline")...)
			err = fmt.Errorf("invalid command")
		}
	}

	result = append(result, []byte(DELIMITER)...)
	return
}

func handlePx(c <-chan time.Time, key string) {
	<-c
	delete(db, key)
}
