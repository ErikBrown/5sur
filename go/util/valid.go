package util

import (
	"errors"
	"net/http"
	"strconv"
)


func ValidCreateSubmit(r *http.Request) (int, int, int, float64, error) {

	//Check if the values that should be ints actually are. If not, return error.
	//Check if values are empty.
	if r.FormValue("Leaving") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
		return 0, 0, 0, 0, errors.New("Please fully fill out the form")
	}
	originId, err := strconv.Atoi(r.FormValue("Origin"))
	if err != nil {
		return 0, 0, 0, 0, errors.New("Invalid origin")
	}

	destinationId, err := strconv.Atoi(r.FormValue("Destination"))
	if err != nil {
		return 0, 0, 0, 0, errors.New("Invalid destination")
	}
	seats, err := strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return 0, 0, 0, 0, errors.New("Invalid number of seats")
	}
	fee, err := strconv.ParseFloat(r.FormValue("Fee"), 64)
	if err != nil {
		return 0, 0, 0, 0, errors.New("Invalid fee")
	}


	// Check if origin and destination are the same
	if r.FormValue("Origin") == r.FormValue("Destination") {
		return 0, 0, 0, 0, errors.New("Please enter different origin and destination")
	}

	if fee > 100 {
		return 0, 0, 0, 0, errors.New("Fee is too high")
	}

	if seats > 8 {
		return 0, 0, 0, 0, errors.New("Too many seats")
	}

	return originId, destinationId, seats, fee, nil

}