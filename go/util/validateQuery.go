package util

import (
	"strconv"
	"net/url"
)

type QueryFields struct {
	Origin int
	Destination int
	Time int
}

type MyError struct {
	What string
}

func (e *MyError) Error() string {
	return e.What
}

func ValidQueryString(u *url.URL) (QueryFields, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		f := QueryFields {0,0,0}
		e := &MyError{"Empty Field"}
		return f, e
	}
	if _,ok := m["o"]; !ok {
		f := QueryFields {0,0,0}
		e := &MyError{"Missing origin"}
		return f, e
	}
	if _,ok := m["d"]; !ok {
		f := QueryFields {0,0,0}
		e := &MyError{"Missing destination"}
		return f, e
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := &MyError{"Origin is not an integer"}
		return f, e
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := &MyError{"Destination is not an integer"}
		// redirect to index to prevent sql injection and end function
		return f, e
	}
	f := QueryFields{city1, city2, 0}
	var e error
	return f, e
}