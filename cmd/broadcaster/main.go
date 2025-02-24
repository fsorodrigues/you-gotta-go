package main

import (
	"log"
	"os"
	"strconv"

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
	ENCODING_VERSION, err := strconv.Atoi(os.Getenv("ENCODING_VERSION"))
	if err != nil {
		log.Fatalln("Error converting ENCODING_VERSION. Is it a valid int?")
	}

	app := Comms{
		API_KEY:          API_KEY,
		BASE_URL:         BASE_URL,
		TCP_PORT:         TCP_PORT,
		ConnectedDevices: make(map[string]Device),
	}

	app.Listen(ENCODING_VERSION)
}
