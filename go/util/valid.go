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
		f := ""
		e := errors.New("Empty Field")
		return f, e
	}
	if _,ok := m["t"]; !ok {
		f := ""
		e := errors.New("Missing token")
		return f, e
	}
	f := m["t"][0]
	return f, nil
}

func ValidCreateSubmit(r *http.Request) (CreateSubmitPost, error) {
	//Check if the values that should be ints actually are. If not, return error.
	//Check if values are empty.
	values := CreateSubmitPost{}
	if r.FormValue("Date") == "" || r.FormValue("Time") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
		return values, errors.New("Please fully fill out the form")
	}
	err := errors.New("")
	values.Origin, err = strconv.Atoi(r.FormValue("Origin"))
	if err != nil {
		return values, errors.New("Invalid origin")
	}

	values.Destination, err = strconv.Atoi(r.FormValue("Destination"))
	if err != nil {
		return values, errors.New("Invalid destination")
	}
	values.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return values, errors.New("Invalid number of seats")
	}
	values.Fee, err = strconv.ParseFloat(r.FormValue("Fee"), 64)
	if err != nil {
		return values, errors.New("Invalid fee")
	}


	// Check if origin and destination are the same
	if r.FormValue("Origin") == r.FormValue("Destination") {
		return values, errors.New("Please enter different origin and destination")
	}

	if values.Fee > 100 {
		return values, errors.New("Fee is too high")
	}

	if values.Seats > 8 {
		return values, errors.New("Too many seats")
	}

	// Date leaving stuff
	timeVar, err := ReturnTime(r.FormValue("Date"), r.FormValue("Time"))
	if err != nil {
		return values, err
	}

	if timeVar.Before(time.Now().Local()) {
		return values, errors.New("Can't make listings in the past, silly")
	}

	if timeVar.After(time.Now().Local().AddDate(0,2,0)) {
		return values, errors.New("Can't make listings this far into the future, silly")
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
		// panic
	}

	m, err := url.ParseQuery(urlParsed.RawQuery)
	if err != nil {
		e := errors.New("Empty Field")
		return ListingQueryFields{}, e
	}
	if _,ok := m["o"]; !ok {
		e := errors.New("Missing origin")
		return ListingQueryFields{}, e
	}
	if _,ok := m["d"]; !ok {
		e := errors.New("Missing destination")
		return ListingQueryFields{}, e
	}
	if _,ok := m["t"]; !ok {
		e := errors.New("Missing date")
		return ListingQueryFields{}, e
	}
	if _,ok := m["h"]; !ok {
		e := errors.New("Missing time")
		return ListingQueryFields{}, e
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		e := errors.New("Origin is not an integer")
		return ListingQueryFields{}, e
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		e := errors.New("Destination is not an integer")
		return ListingQueryFields{}, e
	}

	timeVar, err := ReturnTime(m["t"][0], m["h"][0])
	if err != nil {
		return ListingQueryFields{}, err
	}

	// THIS WILL CHANGE A BIT ONCE WE HAVE THE HH:MM FILTER IN PLACE
	if timeVar.Before(time.Now().Local().Add(time.Minute*-2)) {
		return ListingQueryFields{}, errors.New("Can't search for listings in the past, silly")
	}

	if timeVar.After(time.Now().Local().AddDate(0,2,0)) {
		return ListingQueryFields{}, errors.New("Listings don't exist this far into the future, silly")
	}

	f := ListingQueryFields{city1, city2, timeVar.Format("2006-01-02"), timeVar.Format("15:04")}
	return f, nil
}

func ValidRegister(r *http.Request) error {
		// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" || r.FormValue("Email") == "" {
		return errors.New("Enter a username and password")
	}

	if utf8.RuneCountInString(r.FormValue("Password")) < 6 {
		return errors.New("Password not long enough")
	}

	if r.FormValue("Password") != r.FormValue("Password2"){
		return errors.New("Enter the same username and password")
	}

	if r.FormValue("Email") != r.FormValue("Email2") {
		return errors.New("Enter identical emails")
	}
	return nil
}

func ValidRegisterPost(r *http.Request) (ReservationPost, error) {
	reservePost:= ReservationPost{}
	err := errors.New("")
	if r.FormValue("Seats") == "" || r.FormValue("Listing") == ""{
		return reservePost, errors.New("Missing required fields")
	}
	
	reservePost.ListingId, err = strconv.Atoi(r.FormValue("Listing"))
	if err != nil {
		return reservePost, errors.New("Invalid Listing")
	}
	
	reservePost.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return reservePost, errors.New("Invalid number of seats")
	}
	return reservePost, nil
}

func ValidReserveURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, errors.New("Url parse error")
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, errors.New("Url parse error")
	}
	if _,ok := m["l"]; !ok {
		return 0, errors.New("Missing listing ID")
	}
	listingId, err := strconv.Atoi(m["l"][0])
	if err != nil {
		return 0, errors.New("Invalid listing ID")
	}
	return listingId, nil
}

func ValidDashQuery(u *url.URL) (int, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		e := errors.New("Empty Field")
		return 0, e
	}
	if _,ok := m["i"]; !ok {
		e := errors.New("Missing token")
		return 0, e
	}
	f := m["i"][0]
	i, err := strconv.Atoi(f)
	if err != nil {
		return 0, err
	}
	return i, nil
}