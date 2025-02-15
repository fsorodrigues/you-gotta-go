package main

import (
	"log"
	"os"

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

	app := Comms{
		API_KEY:  API_KEY,
		BASE_URL: BASE_URL,
		TCP_PORT: "1108",
	}

	app.Listen()
}
