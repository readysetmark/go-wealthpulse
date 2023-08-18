package main

import (
	"fmt"
	"os"

	"github.com/readysetmark/go-wealthpulse/pkg/parse"
)

func main() {
	pricesFilePath := os.Getenv("WEALTH_PULSE_PRICES_FILE")
	fmt.Printf("Reading prices from %s\n", pricesFilePath)

	priceBytes, err := os.ReadFile(pricesFilePath)
	if err != nil {
		fmt.Printf("Error reading prices file: %v\n", err)
		os.Exit(1)
	}
	prices, err := parse.ParsePriceDB(string(priceBytes))
	if err != nil {
		fmt.Printf("Error parsing prices file: %v\n", err)
	}
	fmt.Printf("Parsed %d prices\n", len(prices))

	fmt.Println("First 5 prices:")
	for i, v := range prices {
		fmt.Printf("%v\n", v)
		if i >= 5 {
			break
		}
	}
}
