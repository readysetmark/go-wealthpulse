package parse_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/readysetmark/go-wealthpulse/pkg/parse"
)

func TestParsePrice(t *testing.T) {
	tests := map[string]parse.Price{
		"P 2021-08-28 \"WP\" $25.4400": {
			Date: time.Date(2021, time.August, 28, 0, 0, 0, 0, time.UTC),
			Unit: "WP",
			Price: parse.Amount{
				Unit:     "$",
				Quantity: "25.4400",
			},
		},
	}

	for test, want := range tests {
		got, err := parse.ParsePrice(test)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestParsePrices(t *testing.T) {
	test := "P 2022-02-20 \"WP\" $25.0000\r\nP 2022-02-21 \"WP\" $25.4400\r\n"

	want := []parse.Price{
		{
			Date: time.Date(2022, time.February, 20, 0, 0, 0, 0, time.UTC),
			Unit: "WP",
			Price: parse.Amount{
				Unit:     "$",
				Quantity: "25.0000",
			},
		},
		{
			Date: time.Date(2022, time.February, 21, 0, 0, 0, 0, time.UTC),
			Unit: "WP",
			Price: parse.Amount{
				Unit:     "$",
				Quantity: "25.4400",
			},
		},
	}

	got, err := parse.ParsePriceDB(test)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
