package main

import (
	"bytes"
	"log"
	"os"
	"you-gotta-go/cmd/broadcaster/messages"

	// "you-gotta-go/cmd/scraper"

	"github.com/joho/godotenv"
)

type Broadcaster interface {
	Listen()
	acceptConnection()
	handleConnection()
}

// type Connection interface {
// 	SendMsg()
// }

// func () SendMsg() {
// }

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	API_KEY := os.Getenv("API_KEY")
	BASE_URL := os.Getenv("BASE_URL")
	TCP_PORT := os.Getenv("TCP_PORT")

	app := Comms{
		API_KEY:        API_KEY,
		BASE_URL:       BASE_URL,
		TCP_PORT:       TCP_PORT,
		HoldingBuffer:  make([]byte, 25),
		IncomingMsg:    messages.MsgBuf{Msg: bytes.Buffer{}, MsgComplete: false},
		BytesAvailable: 0,
		BytesRead:      0,
		Reading:        false,
	}

	app.Listen()
}
