package util

import (
	"errors"
)


func ValidCreateSubmid(r *http.Request) error {

	//Check if the values that should be ints actually are. If not, return error.

	originId, err := strconv.Atoi(r.FormValue("Origin"))
	if err != nil {
		return errors.New("Invalid origin")
	}

	destinationId, err := strconv.Atoi(r.FormValue("Destination"))
	if err != nil {
		return errors.New("Invalid destination")
	}
	seats, err := strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return errors.New("Invalid number of seats"))
	}
	fee, err := strconv.ParseFloat(r.FormValue("Fee"), 64)
	if err != nil {
		return errors.New("Invalid fee")
	}

	//Check if values are empty.
	if r.FormValue("Leaving") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
		return errors.New("Please fully fill out the form")
	}
	// Check if origin and destination are the same
	if r.FormValue("Origin") == r.FormValue("Destination") {
		return errors.New("Please enter different origin and destination")
	}

	if fee > 100 {
		return errors.New("Fee is too high")
	}

	if seats > 8 {
		return errors.New("Too many seats")
	}

	return nil

}