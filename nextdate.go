package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}

	dateFormat, err := time.Parse(DateStyle, date)
	if err != nil {
		return "", err
	}

	repeatParts := strings.Split(repeat, " ")

	switch repeatParts[0] {
	case "d":
		if len(repeatParts) == 1 {
			return "", errors.New("repeat without day param")
		}

		day, err := strconv.Atoi(repeatParts[1])
		if err != nil {
			return "", errors.New("repeat with uncorect day format")
		}

		if day > 400 {
			return "", errors.New("repeat with more than MAX day")
		}

		returnDate := dateFormat.AddDate(0, 0, day)

		for returnDate.Before(now) {
			returnDate = returnDate.AddDate(0, 0, day)
		}

		return returnDate.Format(DateStyle), nil

	case "y":
		returnDate := dateFormat.AddDate(1, 0, 0)

		for returnDate.Before(now) {
			returnDate = returnDate.AddDate(1, 0, 0)
		}

		return returnDate.Format(DateStyle), nil

	case "w":
		return "", errors.New("repeat is unknown rule")
	case "m":
		return "", errors.New("repeat is unknown rule")
	default:
		return "", errors.New("repeat is unknown rule")
	}
}
