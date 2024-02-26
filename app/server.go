package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const PORT = 6379

var (
	db = map[string]string{}
	fs = flag.NewFlagSet("fs", flag.ContinueOnError)
)

func init() {
	fDir := fs.String("dir", "", "persistence data dir")
	fDbFilename := fs.String("dbfilename", "", "persistence data file name")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing error: %v\n", err.Error())
	}

	if *fDir != "" && *fDbFilename != "" {
		file, err := os.Open(path.Join(*fDir, *fDbFilename))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		r := bufio.NewReader(file)
		lines, err := r.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("---file---\n%v", string(lines))
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
		buf := make([]byte, 512)
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
		response, err := handleMessage(buf[:n])
		if err != nil {
			log.Printf("message handling error : %v\n", err.Error())
		}

		_, err = conn.Write(response.Bytes())
		if err != nil {
			log.Println("write error: ", err.Error())
		}
	}
}

func handleMessage(msg []byte) (*Resp, error) {
	incoming := NewResp(msg)

	result := NewResp([]byte{})
	var err error

	switch incoming.Type {
	case Array:
		switch strings.ToLower(incoming.Value[1]) {
		case "ping":
			result.Type = String
			result.SetPong()

		case "echo":
			result.Type = Bulk
			result.SetValue(incoming.Value[3])

		case "get":
			result.Type = Bulk
			value, ok := db[incoming.Value[3]]
			if ok {
				result.SetValue(value)
			} else {
				result.Type = Bulk
				result.Value = []string{"-1"}
				err = fmt.Errorf("key not found")
			}

		case "set":
			key, value := incoming.Value[3], incoming.Value[5]
			db[key] = value

			if incoming.Length > 3 && strings.ToLower(incoming.Value[7]) == "px" {
				duration, err := strconv.Atoi(incoming.Value[9])
				if err != nil {
					return nil, err
				}
				// delete after px
				go func(c <-chan time.Time, key string) {
					<-c
					delete(db, key)
				}(time.After(time.Millisecond*time.Duration(duration)), key)
			}
			result.SetOK()

		case "config":
			switch strings.ToLower(incoming.Value[3]) {
			case "get":
				rflag := fs.Lookup(incoming.Value[5])
				if rflag != nil {
					result.Type = Array
					value := rflag.Value.String()
					if value == "" {
						value = rflag.DefValue
					}
					result.AppendBulk(incoming.Value[5], value)
				} else {
					result.Type = Bulk
					result.Value = []string{"-1"}
					err = fmt.Errorf("config not found")
				}
			case "set":
				flagName := incoming.Value[5]
				newValue := incoming.Value[7]
				rflag := fs.Lookup(flagName)
				if rflag != nil {
					rflag.Value.Set(newValue)
					result.SetOK()
				} else {
					result.Type = Bulk
					result.Value = []string{"-1"}
					err = fmt.Errorf("config not found")
				}
			}

		default:
			result.Type = String
			result.SetValue("hi marceline")
			err = fmt.Errorf("invalid command")
		}
	}

	if err2 := result.Parse(); err2 != nil {
		err = err2
	}

	return result, err
}
