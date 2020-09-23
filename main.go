package main

import (
	"encoding/json"
	"errors"
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
	fmt.Println("Analyzing stocks...")

	if err := verifyArgs(os.Args[1:]); err != nil {
		fmt.Println("invalid input:", err)
	}

	amt, err := getTodayGainLoss(os.Args[1:])
	if err != nil {
		fmt.Println("could not retrieve stock information:", err)
	}

	if isGreenDay(amt) {
		fmt.Println("Your stocks are up!")
	} else {
		fmt.Println("Your stocks have seen better days")
	}
}

func verifyArgs(args[] string) error {
	if len(args) % 2 != 0 {
		return errors.New("expected space separated list of {ticker symbol} {count}")
	}
	for i := 0; i < len(args); i = i+2 {
		if !IsLetter(args[i]) {
			return errors.New("ticker symbol must only contain letters but was " + args[i])
		}
	}
	for i := 1; i < len(args); i = i+2 {
		if _, err := strconv.Atoi(args[i]); err != nil {
			return errors.New("stock count must be integer but was " + args[i])
		}
	}
	return nil
}

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isGreenDay(amt *amount) bool {
	if amt.dollar >= 0 && amt.cent >= 0 {
		return true
	}
	return false
}
type stock struct {
	Country string `json:country`
	Symbol string `json:symbol`
	Price string `json:price`
}
type respBody struct {
	Response []stock `json:response`
}

func getStock(ticker string, wg *sync.WaitGroup, c chan<- amount) {
	defer wg.Done()
	fmt.Println("checking price " + ticker)

	// get stock price for the day
	url := fmt.Sprintf("https://fcsapi.com/api-v2/stock/latest?symbol=%s&access_key=%s",ticker, "JWtSLcs045NL95a6GhHs6oYvt46dzbq3EBsPQXIiA8bLrBfUwC")
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error reading stock", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	respBody1 := respBody{}
	jsonErr := json.Unmarshal(body, &respBody1)

	if jsonErr != nil {
		log.Fatal(jsonErr)
		c <- amount{0, 0} // fixme
	}

	for _, st := range respBody1.Response {
		if st.Country == "united-states" {
			fmt.Println(st.Country)
			fmt.Println(st.Symbol)
			fmt.Println(st.Price)
			amtArr := strings.Split(st.Price, ".")
			dollar, err := strconv.Atoi(amtArr[0])
			cent, err := strconv.Atoi(amtArr[1])
			if err != nil {
				log.Fatal(err)
				c <- amount{0, 0} // fixme
			}
			c <- amount{dollar, cent}
		}
	}
}

func getTodayGainLoss(args[] string) (*amount, error) {
	var wg sync.WaitGroup
	prices := make(chan amount, len(args)/2)

	for i := 0; i < len(args); i = i+2 {
		wg.Add(1)
		go getStock(args[i], &wg, prices)
	}

	wg.Wait()
	close(prices)

	runningTotal := amount{0, 0}
	for amt := range prices {
		runningTotal.dollar += amt.dollar
		runningTotal.cent += amt.cent
	}
	return &runningTotal, nil
}
