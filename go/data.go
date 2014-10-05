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

func customError() string{
	output := `<!doctype html>

<html>

<head>

<meta charset='utf-8'>

<link rel="stylesheet" type="text/css" href="http://192.241.219.35/404_style.css" />
<link href='http://fonts.googleapis.com/css?family=Exo:700' rel='stylesheet' type='text/css'>
<link href='http://fonts.googleapis.com/css?family=Open+Sans' rel='stylesheet' type='text/css'>
<link rel="icon" type="image/png" href="favicon.ico">

<title>RideShare App</title>

</head>

<body>
<div id="center">
	<h1>404 Error</h1>
	<span><a href="http://192.241.219.35/">Return to homepage</a></span>
</div>
</body>

</html>`
	return output
}

func showListings(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("Name")
	fmt.Fprint(w, name)
	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		fmt.Fprint(w, customError())
		return
	}
	if _,ok := m["o"]; !ok {
		fmt.Fprint(w, customError())
		return
	}
	if _,ok := m["d"]; !ok {
		fmt.Fprint(w, customError())
		return
	}
	m1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		fmt.Fprint(w, customError())
		return
	}
	m2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		fmt.Fprint(w, customError())
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
	<div id="header">
		<h1>RideChile</h1>
		<ul id="account_nav">
			<li>UserName</li>
			<li>MsgIcon</li>
			<li>Logout</li>
		</ul>
	</div>

	<div id="search_wrapper">
		<form method="post" action="http://192.241.219.35/go/l/">
			<select name="Origin">
				<option value="1">City 1</option>
				<option value="2">City 2</option>
				<option value="3">City 3</option>
				<option value="4">City 4</option>
			</select>
			To
			<select name="Destination">
				<option value="1">City 1</option>
				<option value="2">City 2</option>
				<option value="3">City 3</option>
				<option value="4">City 4</option>
			</select>
			<input type="submit" value="Go">
		</form>
	</div>
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