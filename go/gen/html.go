package gen

import (
	"strconv"
	"data/util"
	"fmt"
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
	Driver int
	Picture string
	DateLeaving string
	Origin string
	Destination string
	Seats int
	Fee float32
}

func HeaderHtml(h *Header) string {
	temp := `<!doctype html>
	<html>
		<head>
			<title>` + h.Title + `</title>
			<link href="https://fonts.googleapis.com/css?family=Montserrat:400,700|Open+Sans:400,400italic,600,300,700,800|Bitter:400,400italic,700" rel="stylesheet" type="text/css">
			<link rel="stylesheet" type="text/css" href="https://192.241.219.35/style.css" />
		</head>
	<body>
	<div id="header">
		<h1><a href="https://192.241.219.35">RideChile</a></h1>
		<ul id="account_nav">`

	if h.User == "" {
		temp += `<li>Login</li>`
	} else {
		temp += `<li>` + h.User + `</li>
		<li>` + strconv.Itoa(h.Messages) + ` Msgs</li>
		<li>Logout</li>`
	}	

	temp += `</ul>
	</div>`

	return temp
}

func FilterHTML(cities []City, o int, d int) string {
	temp := `
			<div id="search_wrapper">
			<form method="post" action="https://192.241.219.35/go/l/">
				<div id="city_wrapper">
				<select name="Origin" id="origin_select" class="city_select">`
	for i := range cities{
		temp += optionHTML(cities[i], o)
	}
	temp += `
			</select>
			<span class="to">&#10132;</span>
			<select name="Destination" id="destination_select" class="city_select">`
	for i := range cities{
		temp += optionHTML(cities[i], d)
	}
	temp += `
			</select>
			</div>
			<input type="submit" value="Go">
		</form>
	</div>`
	return temp
}

func optionHTML(c City, i int) string {
	selected := ""
	if c.Id == i {
		selected = " selected"
	}
	return `
	<option value=` + strconv.Itoa(c.Id) + selected + `>` + c.Name + `</option>`
}

func ListingsHTML(l []Listing) string{
	output := ""
	for i := range l {
	date := util.CustomDate(l[i].DateLeaving)
	output += `
	<ul class="list_item">
		<li class="listing_user">
			<img src="https://192.241.219.35/` + l[i].Picture + `" alt="User Picture">
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
		</li>
	</ul>
	`
	}
	return output
}

func FooterHtml() string{
return `</body>
</html>`
}

func Error404() string{
	return `
	<!doctype html>

	<html>

	<head>

	<meta charset='utf-8'>

	<link rel="stylesheet" type="text/css" href="https://192.241.219.35/404_style.css" />
	<link href='https://fonts.googleapis.com/css?family=Exo:700' rel='stylesheet' type='text/css'>
	<link href='https://fonts.googleapis.com/css?family=Open+Sans' rel='stylesheet' type='text/css'>

	<title>RideShare App</title>

	</head>

	<body>
	<div id="center">
		<h1>404 Error</h1>
		<span><a href="https://192.241.219.35/">Return to homepage</a></span>
	</div>
	</body>

	</html>
	`
}