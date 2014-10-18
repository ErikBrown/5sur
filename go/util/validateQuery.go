package util

import (
	"strconv"
	"net/url"
	"errors"
)

type QueryFields struct {
	Origin int
	Destination int
	Time int
}

func ValidQueryString(u *url.URL) (QueryFields, error) {
	// ParseQuery parses the URL-encoded query string and returns a map listing the values specified for each key.
	// ParseQuery always returns a non-nil map containing all the valid query parameters found
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		f := QueryFields {0,0,0}
		e := errors.New("Empty Field")
		return f, e
	}
	if _,ok := m["o"]; !ok {
		f := QueryFields {0,0,0}
		e := errors.New("Missing origin")
		return f, e
	}
	if _,ok := m["d"]; !ok {
		f := QueryFields {0,0,0}
		e := errors.New("Missing destination")
		return f, e
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := errors.New("Origin is not an integer")
		return f, e
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := errors.New("Destination is not an integer")
		// redirect to index to prevent sql injection and end function
		return f, e
	}
	f := QueryFields{city1, city2, 0}
	return f, nil
}