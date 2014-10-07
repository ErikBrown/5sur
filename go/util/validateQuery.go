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

func ValidPost(origin string, destination string) (QueryFields, error) {
	if len(origin) > 0 || len(destination) > 0 {
		city1, err := strconv.Atoi(origin)
		if err != nil{
			f := QueryFields {0,0,0}
			e := &MyError{"Errors!"}
			return f, e
		}
		city2, err := strconv.Atoi(destination)
		if err != nil{
			f := QueryFields {0,0,0}
			e := &MyError{"Errors!"}
			return f, e
		}
		f := QueryFields{city1, city2, 0}
		var e error
		return f, e
	} else {
		f := QueryFields {0,0,0}
		var e error
		return f, e
	}
}

func ValidQueryString(u *url.URL) (QueryFields, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		f := QueryFields {0,0,0}
		e := &MyError{"Errors!"}
		return f, e
	}
	if _,ok := m["o"]; !ok {
		f := QueryFields {0,0,0}
		e := &MyError{"Errors!"}
		return f, e
	}
	if _,ok := m["d"]; !ok {
		f := QueryFields {0,0,0}
		e := &MyError{"Errors!"}
		return f, e
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := &MyError{"Errors!"}
		return f, e
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		f := QueryFields {0,0,0}
		e := &MyError{"Errors!"}
		// redirect to index to prevent sql injection and end function
		return f, e
	}
	f := QueryFields{city1, city2, 0}
	var e error
	return f, e
}