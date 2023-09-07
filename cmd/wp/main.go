package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/readysetmark/go-wealthpulse/internal/parse"
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
	fmt.Printf("Read %d prices\n", len(prices))

	fmt.Println("Scraping prices for symbols:")
	// TODO: Ripe for parallelization!
	for symbol, code := range fundDataCodeMap {
		fmt.Printf("symbol: %s, code: %s\n", symbol, code)
		retrievedPrices, err := getPricesForSymbol(symbol, code)
		if err != nil {
			fmt.Printf("Error getting prices: %v", err)
			continue
		}

		// Add new prices
		latestPrice := latestPriceForSymbol(symbol, prices)
		// fmt.Printf("Latest price found: %v\n", latestPrice)
		for _, retrievedPrice := range retrievedPrices {
			if retrievedPrice.Date.After(latestPrice.Date) {
				fmt.Printf("Adding price: %s\n", retrievedPrice)
				prices = append(prices, retrievedPrice)
			}
		}
	}

	// Sort prices (by unit, then date)
	sort.Slice(prices, func(i, j int) bool {
		return (prices[i].Symbol < prices[j].Symbol) || (prices[i].Symbol == prices[j].Symbol && prices[i].Date.Before(prices[j].Date))
	})

	// Write prices to output file
	const outputFileName = "temp_prices.txt"
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
	}
	defer outputFile.Close()

	for _, price := range prices {
		outputFile.Write([]byte(fmt.Sprintf("%s\r\n", price)))
	}
	fmt.Printf("Wrote prices to %s\n", outputFileName)
}

func latestPriceForSymbol(symbol string, prices []parse.Price) parse.Price {
	var latestPrice parse.Price
	for _, price := range prices {
		if price.Symbol == symbol {
			if (latestPrice.Symbol == "") || (latestPrice.Symbol != "" && price.Date.After(latestPrice.Date)) {
				latestPrice = price
			}
		}
	}
	return latestPrice
}

func getPricesForSymbol(symbol string, code string) ([]parse.Price, error) {
	// TODO: clean this URL up so it is legible (i.e. break out fields, format and URL encode them)
	url := "TODO: FUND DATA URL"

	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println("Fund Data Response:")
	// fmt.Println(string(body))

	var fundDataResponse FundDataResponse
	err = json.Unmarshal(body, &fundDataResponse)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Data points received: %d\n", len(fundDataResponse.ChartData[0][0].RawData))
	// fmt.Printf("First price received: %s\n", toPrice(symbol, fundDataResponse.ChartData[0][0].RawData[0]))
	// fmt.Printf("Last price received: %s\n", toPrice(symbol, fundDataResponse.ChartData[0][0].RawData[len(fundDataResponse.ChartData[0][0].RawData)-1]))

	if len(fundDataResponse.ChartData) == 0 || len(fundDataResponse.ChartData[0]) == 0 {
		return nil, errors.New("unexpected response: chart data not found")
	}
	prices := make([]parse.Price, 0)
	for _, rawData := range fundDataResponse.ChartData[0][0].RawData {
		prices = append(prices, toPrice(symbol, rawData))
	}

	return prices, nil
}

type FundDataResponse struct {
	ChartData [][]ChartData `json:"chart_data"`
}

// expect 2 values in the RawData inner array:
//  1. unix timestamp in milliseconds (treat as UTC)
//  2. the price
type ChartData struct {
	RawData [][]float64 `json:"raw_data"`
}

func toPrice(symbol string, rawFundPrice []float64) parse.Price {
	// TODO: len(rawFundPrice) should be 2 or error
	return parse.Price{
		Date:   time.UnixMilli(int64(rawFundPrice[0])).UTC(),
		Symbol: symbol,
		Price: parse.Amount{
			Symbol:   "$",
			Quantity: fmt.Sprintf("%.2f", rawFundPrice[1]), // TODO: Decimal or format properly!
		},
	}
}
