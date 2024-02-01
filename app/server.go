package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp4", "0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
	}
	log.Println("listening on port 6379")
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		log.Fatal("Error accepting connection: ", err.Error())
	}
	defer conn.Close()

	_, err = conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		log.Fatal("Error writing data to connection: ", err.Error())
	}
}
