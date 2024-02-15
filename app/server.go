package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	PORT      = 6379
	DELIMITER = "\r\n"
)

var (
	db map[string]string = map[string]string{}
	fs                   = flag.NewFlagSet("f1", flag.ContinueOnError)
)

func init() {
	fs.String("dir", "/tmp/redis-files", "persistence data dir")
	fs.String("dbfilename", "dump.rdb", "persistence data file name")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("flag parsing error: %v\n", err.Error())
	}
}

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
			log.Printf("message handling error : %v\n", err.Error())
		}

		fmt.Printf("resp: %v\n", string(resp))
		_, err = conn.Write(resp)
		if err != nil {
			log.Println("write error: ", err.Error())
		}
	}
}

func handleMessage(msg []byte) ([]byte, error) {
	splittedMsg := strings.Split(string(msg), DELIMITER)

	result := string(String)
	var err error

	respType := RespT(splittedMsg[0][0])
	switch respType {
	case Array:
		arrayLen, _ := strconv.Atoi(splittedMsg[0][1:])
		switch strings.ToLower(splittedMsg[2]) {
		case "echo":
			for i := 4; i < len(splittedMsg); i += 2 {
				result += splittedMsg[i]
			}

		case "ping":
			result += "PONG"

		case "set":
			key, value := splittedMsg[4], splittedMsg[6]
			db[key] = value

			if arrayLen > 3 && strings.ToLower(splittedMsg[8]) == "px" {
				duration, err := strconv.Atoi(splittedMsg[10])
				if err != nil {
					return nil, err
				}
				// delete after px
				go func(c <-chan time.Time, key string) {
					<-c
					delete(db, key)
				}(time.After(time.Millisecond*time.Duration(duration)), key)
			}

			result += "OK"

		case "get":
			value, ok := db[splittedMsg[4]]
			if ok {
				result += value
			} else {
				result = "$-1"
				err = fmt.Errorf("key not found")
			}

		case "config":
			if strings.ToLower(splittedMsg[6]) == "get" {
				rflag := fs.Lookup(splittedMsg[8])
				if rflag != nil {
					// TODO
					// result = string(Array) + "2" + DELIMITER + string(Bulk) + len(rflag.Name)
				} else {
					result = "$-1"
					err = fmt.Errorf("config not found")
				}

			}

		default:
			result += "hi marceline"
			err = fmt.Errorf("invalid command")
		}
	}

	result += DELIMITER
	return []byte(result), err
}
