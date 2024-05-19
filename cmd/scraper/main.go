package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func getData(url string, api_key string) http.Response {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Error creating HTTP request:", err)
	}
	req.Header.Add("x-api-key", api_key)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	return *resp
}

func parseData(resp http.Response) string {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	API_KEY := os.Getenv("API_KEY")
	BASE_URL := os.Getenv("BASE_URL")
	station := os.Args[1]
	resp := getData(fmt.Sprintf("%s/stop-predictions?stop_id=%s", BASE_URL, station), API_KEY)

	data := parseData(resp)

	fmt.Println(data)
}
