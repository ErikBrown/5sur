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
		f := ListingQueryFields {0,0,""}
		e := errors.New("Empty Field")
		return f, e
	}
	if _,ok := m["o"]; !ok {
		f := ListingQueryFields {0,0,""}
		e := errors.New("Missing origin")
		return f, e
	}
	if _,ok := m["d"]; !ok {
		f := ListingQueryFields {0,0,""}
		e := errors.New("Missing destination")
		return f, e
	}
	if _,ok := m["t"]; !ok {
		f := ListingQueryFields {0,0,""}
		e := errors.New("Missing time")
		return f, e
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		f := ListingQueryFields {0,0,""}
		e := errors.New("Origin is not an integer")
		return f, e
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		f := ListingQueryFields {0,0,""}
		e := errors.New("Destination is not an integer")
		// redirect to index to prevent sql injection and end function
		return f, e
	}
	f := ListingQueryFields{city1, city2, m["t"][0]}
	return f, nil
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

func ValidCreateSubmit(r *http.Request) (CreateSubmitPost, error) {
	//Check if the values that should be ints actually are. If not, return error.
	//Check if values are empty.
	values := CreateSubmitPost{}
	if r.FormValue("Leaving") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
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
	values.Date = ConvertDate(r.FormValue("Leaving"))
	err = CompareDate(values.Date, time.Now().Local().Format(time.RFC3339))
	if err != nil {
		return values, err
	}

	return values, nil
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

func ValidReserveURL(r *http.Request) (string, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return "", errors.New("Url parse error")
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", errors.New("Url parse error")
	}
	if _,ok := m["l"]; !ok {
		return "", errors.New("Missing listing ID")
	}
	return m["l"][0], nil
}