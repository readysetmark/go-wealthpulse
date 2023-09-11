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

// TODO: to configuration
var fundDataCodeMap = map[string]string{
	"TDB900": "TDB900.TO",
	"TDB902": "TDB902.TO",
	"TDB909": "TDB909.TO",
	"TDB911": "TDB911.TO",
}

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

	// Scrape for new prices
	// TODO: Ripe for parallelization!
	foundNew := false
	for symbol, code := range fundDataCodeMap {
		fmt.Printf("Retrieving prices for symbol: %s\n", symbol)
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
				foundNew = true
			}
		}
	}

	if !foundNew {
		fmt.Println("No new prices found")
		return
	}

	// Sort prices (by unit, then date)
	sort.Slice(prices, func(i, j int) bool {
		return (prices[i].Symbol < prices[j].Symbol) || (prices[i].Symbol == prices[j].Symbol && prices[i].Date.Before(prices[j].Date))
	})

	// Write prices to output file
	// const outputFileName = "temp_prices.txt"
	writeFilePath := pricesFilePath + "_write"
	err = writePrices(writeFilePath, prices)
	if err != nil {
		fmt.Printf("error writing prices: %v\n", err)
		return
	}

	err = os.Remove(pricesFilePath)
	if err != nil {
		fmt.Printf("error deleting price file: %v\n", err)
	}

	err = os.Rename(writeFilePath, pricesFilePath)
	if err != nil {
		fmt.Printf("error renaming price file: %v\n", err)
	}
	fmt.Printf("Wrote %d prices to %s\n", len(prices), pricesFilePath)
}

func writePrices(fileName string, prices []parse.Price) error {
	outputFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer outputFile.Close()
	for _, price := range prices {
		outputFile.Write([]byte(fmt.Sprintf("%s\r\n", price)))
	}
	return nil
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
	url := "https://ycharts.com/charts/fund_data.json?securities=id%3AM%3A" + code + "%2Cinclude%3Atrue%2C%2C&calcs=id%3Aprice%2Cinclude%3Atrue%2C%2C&correlations=&format=real&recessions=false&zoom=5&startDate=&endDate=&chartView=&splitType=&scaleType=&note=&title=&source=&units=&quoteLegend=&partner=&quotes=&legendOnChart=&securitylistSecurityId=&clientGroupLogoUrl=&displayTicker=&ychartsLogo=&useEstimates="

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
