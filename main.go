package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"
)

func main() {
	fmt.Println("Analyzing stocks...")

	if err := verifyArgs(os.Args[1:]); err != nil {
		fmt.Println("invalid input:", err)
	}

	dollar, cents, err := getTodayGainLoss(os.Args[1:])
	if err != nil {
		fmt.Println("could not retrieve stock information:", err)
	}

	if isGreenDay(dollar, cents) {
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

func isGreenDay(dollar, cents int) bool {
	if dollar >= 0 && cents >= 0 {
		return true
	}
	return false
}

func getTodayGainLoss(args[] string) (int, int, error) {
	// use int for cents
	return 0, 0, nil
}
