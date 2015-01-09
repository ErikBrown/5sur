package util

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"
	"time"
)

type ListingQueryFields struct {
	Origin int
	Destination int
	Date string
	Time string
}

type CreateSubmitPost struct {
	Origin int
	Destination int
	Seats int
	Fee float64
	Date string
}

type ReservationPost struct {
	ListingId int
	Seats int
}

func ValidAuthQuery(u *url.URL) (string, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", NewError(nil, "Authentication failed", 400)
	}
	if _,ok := m["t"]; !ok {
		return "", NewError(nil, "Authentication failed", 400)
	}
	f := m["t"][0]
	return f, nil
}

func ValidCreateSubmit(r *http.Request) (CreateSubmitPost, error) {
	//if the values that should be ints actually are. If not, return error.
	//Check if values are empty.
	values := CreateSubmitPost{}
	if r.FormValue("Date") == "" || r.FormValue("Time") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
		return values, NewError(nil, "Please fully fill out the form", 400)
	}
	err := errors.New("")
	values.Origin, err = strconv.Atoi(r.FormValue("Origin"))
	if err != nil {
		return values, NewError(nil, "Invalid origin", 400)
	}

	values.Destination, err = strconv.Atoi(r.FormValue("Destination"))
	if err != nil {
		return values, NewError(nil, "Invalid destination", 400)
	}
	values.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return values, NewError(nil, "Invalid number of seats", 400)
	}
	values.Fee, err = strconv.ParseFloat(r.FormValue("Fee"), 64)
	if err != nil {
		return values, NewError(nil, "Invalid fee", 400)
	}

	if r.FormValue("Origin") == r.FormValue("Destination") {
		return values, NewError(nil, "Please enter different origin and destination", 400)
	}

	if values.Fee > 100 {
		return values, NewError(nil, "Fee is too high. Max fee is $100", 400)
	}

	if values.Seats > 8 {
		return values, errors.New("Too many seats. You can only select up to 8 seats")
	}

	// Date leaving stuff
	timeVar, err := ReturnTime(r.FormValue("Date"), r.FormValue("Time"))
	if err != nil {
		return values, NewError(nil, "Invalid date leaving", 400)
	}

	if timeVar.Before(time.Now().Local()) {
		return values, NewError(nil, "Can't make listings in the past", 400)
	}

	if timeVar.After(time.Now().Local().AddDate(0,2,0)) {
		return values, NewError(nil, "Can't make listings more than two months in the future", 400)
	}

	values.Date = timeVar.Format("2006-01-02 15:04:05")

	return values, nil
}

// CHANGE TO var := struct{}
func ValidListingQuery(u *url.URL) (ListingQueryFields, error) {
	// ParseQuery parses the URL-encoded query string and returns a map listing the values specified for each key.
	// ParseQuery always returns a non-nil map containing all the valid query parameters found
	urlParsed, err := url.Parse(u.String())
	if err != nil {
		return ListingQueryFields{}, NewError(nil, "Incorrect url", 400)
	}

	m, err := url.ParseQuery(urlParsed.RawQuery)
	if err != nil {
		return ListingQueryFields{}, NewError(nil, "Missing parameters", 400)
	}
	if _,ok := m["o"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Missing origin", 400)
	}
	if _,ok := m["d"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Missing destination", 400)
	}
	if _,ok := m["t"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Missing date", 400)
	}
	if _,ok := m["h"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Missing time", 400)
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		return ListingQueryFields{}, NewError(nil, "Invalid origin", 400)
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		return ListingQueryFields{}, NewError(nil, "Invalid Destination", 400)
	}

	timeVar, err := ReturnTime(m["t"][0], m["h"][0])
	if err != nil {
		return ListingQueryFields{}, NewError(nil, "Invalid time", 400)
	}

	return ListingQueryFields{city1, city2, timeVar.Format("2006-01-02"), timeVar.Format("15:04")}, nil
}

func ValidRegister(r *http.Request) error {
		// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" || r.FormValue("Email") == "" {
		return NewError(nil, "Fill out all fields", 400)
	}

	if utf8.RuneCountInString(r.FormValue("Password")) < 6 {
		return NewError(nil, "Password must be at least six characters", 400)
	}

	if r.FormValue("Password") != r.FormValue("Password2"){
		return NewError(nil, "Passwords do not match", 400)
	}

	if r.FormValue("Email") != r.FormValue("Email2") {
		return NewError(nil, "Emails do not match", 400)
	}
	return nil
}

func ValidRegisterPost(r *http.Request) (ReservationPost, error) {
	reservePost := ReservationPost{}
	err := errors.New("")
	if r.FormValue("Seats") == "" || r.FormValue("Listing") == ""{
		return ReservationPost{}, NewError(nil, "Missing required fields", 400)
	}
	
	reservePost.ListingId, err = strconv.Atoi(r.FormValue("Listing"))
	if err != nil {
		return ReservationPost{}, NewError(nil, "Invalid Listing", 400)
	}
	
	reservePost.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return ReservationPost{}, NewError(nil, "Invalid number of seats", 400)
	}
	return reservePost, nil
}

func ValidReserveURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, NewError(err, "Internal server error", 500)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(err, "Internal server error", 500)
	}
	if _,ok := m["l"]; !ok {
		return 0, NewError(nil, "Missing listing id", 400)
	}
	listingId, err := strconv.Atoi(m["l"][0])
	if err != nil {
		return 0, NewError(nil, "Invalid listing id", 400)
	}
	return listingId, nil
}

func ValidDashQuery(u *url.URL) (int, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(nil, "Empty query string", 400)
	}
	if _,ok := m["i"]; !ok {
		return 0, NewError(nil, "Invalid url", 400)
	}
	f := m["i"][0]
	i, err := strconv.Atoi(f)
	if err != nil {
		return 0, NewError(nil, "Invalid url", 400)
	}
	return i, nil
}

func ValidMessageURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, NewError(err, "Internal server error", 500)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(err, "Internal server error", 500)
	}
	if _,ok := m["i"]; !ok {
		return 0, NewError(nil, "Missing message recipient", 400)
	}
	userId, err := strconv.Atoi(m["i"][0])
	if err != nil {
		return 0, NewError(nil, "Invalid message recipient", 400)
	}
	return userId, nil
}

func ValidMessagePost(r *http.Request) (int, string, error) {
	if r.FormValue("Recipient") == "" || r.FormValue("Message") == ""{
		return 0, "", NewError(nil, "Missing required fields", 400)
	}
	
	recipient, err := strconv.Atoi(r.FormValue("Recipient"))
	if err != nil {
		return 0, "", NewError(nil, "Invalid message recipient", 400)
	}
	if utf8.RuneCountInString(r.FormValue("Message")) > 500 {
		return 0, "", NewError(nil, "Message too long (500 character max length)", 400)
	}
	return recipient, r.FormValue("Message"), nil
}