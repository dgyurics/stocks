package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
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

func getStock(ticker string, wg *sync.WaitGroup, c chan<- amount) {
	defer wg.Done()
	fmt.Println("checking price " + ticker)
	time.Sleep(time.Second * 2)
	c <- amount{10, 20}
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
