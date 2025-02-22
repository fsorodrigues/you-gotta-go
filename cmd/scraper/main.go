package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func createUrl(BASE_URL string, stop string) string {
	return fmt.Sprintf("%s/stop-predictions?stop_id=%s", BASE_URL, stop)
}

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

func Scrape(BASE_URL string, stop string, API_KEY string) string {
	url := createUrl(BASE_URL, stop)

	resp := getData(url, API_KEY)

	return parseData(resp)
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	API_KEY := os.Getenv("API_KEY")
	BASE_URL := os.Getenv("BASE_URL")
	stop := os.Args[1]

	data := Scrape(BASE_URL, stop, API_KEY)

	fmt.Println(data)
}
