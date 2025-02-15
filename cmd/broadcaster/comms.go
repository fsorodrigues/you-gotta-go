package main

import (
	"fmt"
	"log"
	"net"
)

type Comms struct {
	API_KEY  string
	BASE_URL string
	TCP_PORT string
}

func (c *Comms) Listen() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", c.TCP_PORT))
	if err != nil {
		log.Fatalln("Can't open TCP connection", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Fatalln("Error accepting incoming TCP connection", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	msg := "hello"

	encodedMsg := []byte(msg)

	conn.Write(encodedMsg)

	conn.Close()
}
