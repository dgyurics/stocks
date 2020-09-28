package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type amount struct {
	dollar, cent int
}

func main() {
	log.Println("Analyzing stocks...")
	validateArgs(os.Args[1:])
	printGrandTotal(os.Args[1:])
}

func validateArgs(args[] string) {
	if len(args) % 2 != 0 {
		log.Fatalln("expected space separated list of ticker and count")
	}
	for i := 0; i < len(args); i = i+2 {
		if !isLetters(args[i]) {
			log.Fatalln("ticker symbol must only contain letters but was " + args[i])
		}
	}
	for i := 1; i < len(args); i = i+2 {
		if _, err := strconv.Atoi(args[i]); err != nil {
			log.Fatalln("stock count must be integer but was " + args[i])
		}
	}
}

func isLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

type stock struct {
	Country string `json:country`
	Symbol 	string `json:symbol`
	Price 	string `json:price`
}

type respBody struct {
	Response []stock `json:response`
	Code     int     `json:code`
	Msg      string  `json:msg`
}

func getStockTotal(ticker string, quantity int, wg *sync.WaitGroup, c chan<- *amount) {
	defer wg.Done()

	const stockUri = "https://fcsapi.com/api-v2/stock/latest?symbol=%s&access_key=%s"
	url := fmt.Sprintf(stockUri, ticker, "JWtSLcs045NL95a6GhHs6oYvt46dzbq3EBsPQXIiA8bLrBfUwC")
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}

	respBody1 := respBody{}
	jsonErr := json.Unmarshal(body, &respBody1)

	if jsonErr != nil {
		log.Printf("Error unmarshaling response body for stock %s", ticker)
		return
	}

	if respBody1.Code != 200 {
		log.Printf("Error fetching stock for %s\n", ticker)
		log.Printf("StatusCode: %d ErrorMsg: %s\n", respBody1.Code, respBody1.Msg)
		return
	}

	const countryUnitedStates = "united-states"
	for _, stockEntry := range respBody1.Response {
		if stockEntry.Country == countryUnitedStates {
			log.Printf("%s is at %s today\n", stockEntry.Symbol, stockEntry.Price)
			amt, err := strToAmt(stockEntry.Price)
			if err != nil {
				log.Printf("Error converting %s stock price %s to type amount", ticker, stockEntry.Price)
				return
			}
			c <- amt.multiply(quantity)
		}
	}
}

// FIXME integer overflow

func (amt amount) add(operand amount) *amount {
	result := amount{amt.dollar + operand.dollar, amt.cent + operand.cent}
	result.dollar = result.dollar + result.cent / 100
	result.cent = result.cent % 100
	return &result
}

func (amt amount) multiply(operand int) *amount {
	result := amount{amt.dollar * operand, amt.cent * operand}
	result.dollar = result.dollar + result.cent / 100
	result.cent = result.cent % 100
	return &result
}

func strToAmt(amt string) (*amount, error) {
	amtArr := strings.Split(amt, ".")
	dollar, err := strconv.Atoi(amtArr[0])
	cent, err := strconv.Atoi(amtArr[1])
	return &amount{dollar, cent}, err
}

func printGrandTotal(args[] string) {
	var wg sync.WaitGroup
	prices := make(chan *amount, len(args)/2)

	for i := 0; i < len(args); i = i+2 {
		wg.Add(1)
		j, _ := strconv.Atoi(args[i+1])
		go getStockTotal(args[i], j, &wg, prices)
	}

	wg.Wait()
	close(prices)

	runningTotal := amount{0, 0}
	for amt := range prices {
		runningTotal = *runningTotal.add(*amt)
	}
	log.Printf("Total assets as of today: %d.%d\n", runningTotal.dollar, runningTotal.cent)
}
