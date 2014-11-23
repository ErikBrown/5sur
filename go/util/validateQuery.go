package util

import (
	"strconv"
	"net/url"
	"errors"
)

type ListingQueryFields struct {
	Origin int
	Destination int
	Time string
}

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