package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ScraperError struct {
	Err error
}

func (d ScraperError) Error() string {
	return fmt.Sprintf("Error while scraping. %v", d.Err.Error())
}

func createUrl(BASE_URL string, stop string) string {
	return fmt.Sprintf("%s/stop-predictions?stop_id=%s", BASE_URL, stop)
}

func getData(url string, api_key string) (http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.Response{}, ScraperError{
			Err: errors.New(fmt.Sprintf("Error creating HTTP request: %v\n", err)),
		}
	}
	req.Header.Add("x-api-key", api_key)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return http.Response{}, ScraperError{
			Err: errors.New(fmt.Sprintf("Error making HTTP request: %v\n", err)),
		}
	}

	return *resp, nil
}

func parseData(resp http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", ScraperError{
			Err: errors.New(fmt.Sprintf("Error making HTTP request: %v\n", err)),
		}
	}

	return string(body), nil
}

func Scrape(BASE_URL string, stop string, API_KEY string) (string, error) {
	url := createUrl(BASE_URL, stop)

	resp, err := getData(url, API_KEY)
	if err != nil {
		return "", err
	}

	return parseData(resp)
}
