package main

import (
	"net/http"
	"net/url"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"data/results"
	"data/util"
	"strconv"
)

func generateListing (myListing results.Listing) string{
	var date util.Date = util.CustomDate(myListing.DateLeaving)
	output := `
	<ul class="list_item">
		<li class="listing_user">
			<img src="http://192.241.219.35/` + myListing.Picture + `" alt="User Picture">
			<span class="positive">+100</span>
		</li>
		<li class="date_leaving">
			<div>
				<span class="month">` + date.Month + `</span>
				<span class="day">` + date.Day + `</span>
				<span class="time">` + date.Time + `</span>
			</div>
		</li>
		<li class="city">
			<span>` + myListing.Origin + `</span>
			<span class="to">&#10132;</span>
			<span>` + myListing.Destination + `</span>
		</li>
		<li class="seats">
			<span>` + fmt.Sprintf("%d", myListing.Seats) + `</span>
		</li>
			<li class="fee"><span>$` + fmt.Sprintf("%.2f", myListing.Fee) + `</span>
		</li>
	</ul>
	`
	return output
}

func showListings(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		//404 error
	}
	m1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		//redirect to index to prevent sql injection and end function.
		return
	}
	m2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		// redirect to index to prevent sql injection and end function
		return
	}
	results := results.ReturnListings(m1, m2) // Make struct to store everything
	myString := `
	<!doctype html>
	<html>
		<head>
			<title>Title</title>
			<link href="http://fonts.googleapis.com/css?family=Montserrat:400,700|Open+Sans:400,400italic,600,300,700,800|Bitter:400,400italic,700" rel="stylesheet" type="text/css">
			<link rel="stylesheet" type="text/css" href="http://192.241.219.35/style.css" />
		</head>
	<body>
	`
	for i := range results{
		myString += generateListing(results[i])
	}
	myString += `
	</body>
	</html>
	`
	fmt.Fprint(w, myString)
}

func main() {
	http.HandleFunc("/go/l/", showListings)
	http.ListenAndServe(":8080", nil)
}