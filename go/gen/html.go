package gen

import (
	"strconv"
	"data/util"
	"fmt"
	"time"
)

type Header struct {
	Title string
	User string
	Messages int
}

type City struct {
	Id int
	Name string
}

type Listing struct {
	Id int
	Driver int
	Picture string
	DateLeaving string
	Origin string
	Destination string
	Seats int
	Fee float32
}

type User struct {
	Name string
	Picture string
	Created string
	RatingPositive int
	RatingNegative int
	RidesTaken int
	RidesGiven int
	FavDestination string
	Comments []struct {
		Rating int
		Message string
	}
}

func HeaderHtml(h *Header) string {
	temp := `<!doctype html>
	<html>
		<head>
			<title>` + h.Title + `</title>
			<link href="https://fonts.googleapis.com/css?family=Montserrat:400,700|Open+Sans:400,400italic,600,300,700,800|Bitter:400,400italic,700" rel="stylesheet" type="text/css">
			<link rel="stylesheet" type="text/css" href="https://5sur.com/style.css" />
		</head>
	<body>
	<div id="header">
		<h1><a href="https://5sur.com/">RideChile</a></h1>
		<ul id="account_nav">`

	if h.User == "" {
		temp += `<li><a href="https://5sur.com/login.html">Login</a></li>`
	} else {
		temp += `<li>` + h.User + `</li>
		<li>` + strconv.Itoa(h.Messages) + ` Msgs</li>
		<li><a href="https://5sur.com/logout">Logout</a></li>`
	}	

	temp += `</ul>
	</div>`

	return temp
}

func FilterHtml(cities []City, query util.ListingQueryFields) string {
	temp := `
		<div id="search_wrapper">
			<form method="POST" action="https://5sur.com/l/" id="search_form">
				<select name="Origin" id="origin_select" class="search_option">`
	if query.Origin == 0 {
		temp += `
					<option disabled selected class="blank_option"></option>`
	}
	for i := range cities{
		temp += optionHtml(cities[i], query.Origin)
	}
	temp += `
				</select>
				<span class="to">&#10132;</span>
				<select name="Destination" id="destination_select" class="search_option">`
	if query.Destination == 0 {
		temp += `
					<option disabled selected class="blank_option"></option>`
	}
	for i := range cities{
		temp += optionHtml(cities[i], query.Destination)
	}
	temp += `
				</select>
				<span class="to">&#128343;</span>
				<input type="text" name="Date" placeholder="Select date..." autocomplete="off" value="`
	// The error for this should have already been checked
	convertedDate, convertedTime, _ := util.ReturnTimeString(true, query.Date, query.Time)
	temp += convertedDate
	temp += `" id="date_box" class="search_option">

				<input type="text" name="Time" placeholder="Select time..." autocomplete="off" value="`
	temp += convertedTime
	temp += `" id="time_box" class="search_option">
				<div id="calendar_wrapper">
					<div id="month_wrapper">
						<span id="month_left">&#9664;</span>
						<span id="month_title"></span>
						<span id="month_right">&#9654;</span>
					</div>
					<table id="calendar">
						<thead>
							<tr>
								<th>lu</th>
								<th>ma</th>
								<th>mi</th>
								<th>ju</th>
								<th>vi</th>
								<th>sa</th>
								<th>su</th>
							</tr>
						</thead>
					</table>
				</div>
				<input type="submit" name="FilterSubmit" value="Find a ride!" id="search_submit">
			</form>
		</div>`
	return temp
}

func optionHtml(c City, i int) string {
	selected := ""
	if c.Id == i {
		selected = " selected"
	}
	return `
					<option value="` + strconv.Itoa(c.Id) + `"` + selected + `>` + c.Name + `</option>`
}

func ListingsHtml(l []Listing) string{
	output := ""
	for i := range l {
	date := util.PrettyDate(l[i].DateLeaving, true)
	output += `
	<ul class="list_item">
		<li class="listing_user">
			<img src="https://5sur.com/` + l[i].Picture + `" alt="User Picture">
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
			<span>` + l[i].Origin + `</span>
			<span class="to">&#10132;</span>
			<span>` + l[i].Destination + `</span>
		</li>
		<li class="seats">
			<span>` + strconv.Itoa(l[i].Seats) + `</span>
		</li>
			<li class="fee"><span>$` + fmt.Sprintf("%.2f", l[i].Fee) + `</span>
			<li class="reserve"><a href="https://5sur.com/reserve?l=` + strconv.Itoa(l[i].Id) + `">Resrv</a></li>
		</li>
	</ul>
	`
	}
	return output
}

func ReserveHtml(l string) string {
	output := `<form method="post" action="https://5sur.com/reserveSubmit" id="reserve_form">
		<span>This input will be hidden:</span>
		<input name="Listing" type="text" id="listing_id_input" value="` + l +`" readonly>
		<br />
		<span>Number of Seats</span>
		<select name="Seats" id="seats_select" class="reserve_input">
			<option selected value="1">1</option>
			<option value="2">2</option>
			<option value="3">3</option>
			<option value="4">4</option>
		</select>
		<br />
		<span>Message</span>
		<input type="text" name="Message" id="message_input" class="reserve_input">
		<br />
		<input type="submit" value="Go">
	</form>`
	return output
}

func CreateListingHtml(u string, c []City) string {
	output := `<form method="post" action="https://5sur.com/createSubmit" id="create_listing_form">
		<span>Date leaving: </span>
		<input name="Date" type="text">
		<span>Time leaving: </span>
		<input name="Time" type="text">
		<br />
		<span>Origin</span>
		<select name="Origin" class="submit_input">
			<option disabled selected class="blank_option"></option>`
		for i := range c{
			output += optionHtml(c[i], 0)
		}
		output += `
		</select>
		<span>Desination</span>
		<select name="Destination" class="submit_input">
			<option disabled selected class="blank_option"></option>`
		for i := range c{
			output += optionHtml(c[i], 0)
		}
		output += `</select>
		<br />
		<span>Seats</span>
		<input type="text" name="Seats" class="submit_input">
		<br />
		<span>Fee</span>
		<input type="text" name="Fee" class="submit_input">
		<br />
		<input type="submit" value="Go">
	</form>`
	return output
}

func DashHtml(dashListing []DashListing, listing SpecificListing) string {
	output := `<div class="sidebar" id="main_sidebar">
	<ul>
		<li class="message_icon">
			<a href="https://5sur.com/dashboard/messages"></a>
		</li>
		<li class="listings_icon selected_dark">
			<a href="https://5sur.com/dashboard/listings"></a>
		</li>
		<li class="reservation_icon">
			<a href="https://5sur.com/dashboard/reservations"></a>
		</li>
		<li class="settings_icon">
			<a href="https://5sur.com/dashboard/settings"></a>
		</li>
	</ul>
</div>

<div class="sidebar" id="sub_sidebar">
	<h2>Listings</h2>
	<ul>`
	for i := range dashListing{
		output += sidebarListing(dashListing[i], listing.ListingId)
	}
	output +=
	`
	</ul>
</div>`

if listing.ListingId == 0 {
	return output
}

output += dashSpecificListing(listing)

return output
}

func sidebarListing(d DashListing, l int) string {
	output := ""
	if d.ListingId == l {
		output += `<li class="selected_light">`
	} else {
		output += `<li>`
	}
	output += `
			<a href="https://5sur.com/dashboard/listings?l=` + strconv.Itoa(d.ListingId) + `">
				<div class="calendar_icon">
					<span class="calendar_icon_month">` + d.Month + `</span>
					<span class="calendar_icon_day">` + d.Day + `</span>
				</div>
				<span class="sidebar_text">` + d.Origin + ` &#10132; ` + d.Destination + `</span>
				`
	if d.Alert {
		output += `<span class="sidebar_alert">!</span>`
	}
	output += `			
			</a>
		</li>`
	return output
}

func dashSpecificListing(l SpecificListing) string {
	output := `<div id="dash_content" class="dash_listings">
	<div id="dash_title">
		<h3>` + l.Origin + ` &#10132; ` + l.Destination + `</h3>
	</div>
	<div id="registered_passengers" class="passengers">
		<span>Registered</span>
		<ul>`
		for i := range l.RegisteredUsers {
			output += dashRegisteredPassenger(l.RegisteredUsers[i], l.ListingId)
		}
		output += `
		</ul>
	</div>
	<div id="pending_passengers" class="passengers">
		<span>Pending</span>
		<ul>`
		for i := range l.PendingUsers {
			output += dashAcceptedPassenger(l.PendingUsers[i], l.ListingId)
		}
		output += `
		</ul>
	</div>
</div>
`
return output
}

func dashRegisteredPassenger(u RegisteredUser, l int) string {
	output := `<li>
				<img src="https://5sur.com/` + u.Picture + `" alt="usr image">
				<span>` + u.Name + `</span>
				<form class="passenger_form" method="POST" action="https://5sur.com/dashboard/listings?l=` + strconv.Itoa(l) + `">
					<input name="asd" value="` + strconv.Itoa(u.Id) + `" id="passenger_reject" type="submit">
					<!--
					<input name="m" value="` + strconv.Itoa(u.Id) + `" id="passenger_message" type="submit">
					-->
				</form>
				<span class="passenger_seats">1 Seat</span>
			</li>`
	return output
}

func dashAcceptedPassenger(u PendingUser, l int) string {
	output := `<li>
				<img src="https://5sur.com/` + u.Picture + `" alt="usr image">
				<span>` + u.Name + `</span>
				<form class="passenger_form" method="POST" action="https://5sur.com/dashboard/listings?l=` + strconv.Itoa(l) + `">
					<input name="a" value="` + strconv.Itoa(u.Id) + `" id="passenger_accept" type="submit">
					<input name="r" value="` + strconv.Itoa(u.Id) + `" id="passenger_reject" type="submit">
					<!--
					<input name="m" value="` + strconv.Itoa(u.Id) + `" id="passenger_message" type="submit">
					-->
				</form>
				<span class="passenger_seats">1 Seat</span>
			</li>`
	return output
}

// Move specific scrips to specific pages!
func FooterHtml() string{
return `
<script src="https://5sur.com/script.js"></script>
</body>
</html>`
}

func Error404() string{
	return `
	<!doctype html>

	<html>

	<head>

	<meta charset='utf-8'>

	<link rel="stylesheet" type="text/css" href="https://5sur.com/404_style.css" />
	<link href='https://fonts.googleapis.com/css?family=Exo:700' rel='stylesheet' type='text/css'>
	<link href='https://fonts.googleapis.com/css?family=Open+Sans' rel='stylesheet' type='text/css'>

	<title>RideShare App</title>

	</head>

	<body>
	<div id="center">
		<h1>404 Error</h1>
		<span><a href="https://5sur.com/">Return to homepage</a></span>
	</div>
	</body>

	</html>
	`
}