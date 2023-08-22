package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/readysetmark/go-wealthpulse/pkg/parse"
)

var fundDataCodeMap = map[string]string{}

func main() {
	pricesFilePath := os.Getenv("WEALTH_PULSE_PRICES_FILE")
	fmt.Printf("Reading prices from %s\n", pricesFilePath)

	pricesFileBytes, err := os.ReadFile(pricesFilePath)
	if err != nil {
		fmt.Printf("Error reading prices file: %v\n", err)
		os.Exit(1)
	}
	prices, err := parse.ParsePriceDB(string(pricesFileBytes))
	if err != nil {
		fmt.Printf("Error parsing prices file: %v\n", err)
	}
	fmt.Printf("Parsed %d prices\n", len(prices))

	fmt.Println("First 5 prices:")
	for i, price := range prices {
		fmt.Printf("%s\n", price)
		if i >= 5 {
			break
		}
	}

	fmt.Println("Scraping prices for symbols:")
	for symbol, code := range fundDataCodeMap {
		fmt.Printf("symbol: %s, code: %s\n", symbol, code)
		err = getPricesForSymbol(symbol, code)
		if err != nil {
			fmt.Printf("Error getting prices: %v", err)
		}
	}
}

func getPricesForSymbol(symbol string, code string) error {
	// TODO: clean this URL up so it is legible (i.e. break out fields, format and URL encode them)
	url := "TODO: FUND DATA URL"

	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// fmt.Println("Fund Data Response:")
	// fmt.Println(string(body))

	var fundDataResponse FundDataResponse
	err = json.Unmarshal(body, &fundDataResponse)
	if err != nil {
		return err
	}
	fmt.Printf("Data points received: %d\n", len(fundDataResponse.ChartData[0][0].RawData))

	return nil
}

type FundDataResponse struct {
	ChartData [][]ChartData `json:"chart_data"`
}

type ChartData struct {
	RawData [][]float64 `json:"raw_data"`
}
