package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/gen"
	"data/util"
	"unicode/utf8"
	"strconv"
)

func ListingsHandler(w http.ResponseWriter, r *http.Request) {
	// POST validation
	if r.FormValue("Origin") != "" && r.FormValue("Destination") != "" {
		http.Redirect(w, r, "https://5sur.com/l/?o=" + r.FormValue("Origin") + "&d=" + r.FormValue("Destination") + "&t=" + util.ConvertDate(r.FormValue("Date")), 301)
		return
	}

	// Query string validation
	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}
	query, err := util.ValidListingQuery(u) // Returns util.QueryFields
	if err != nil {
		// INCORRECT QUERY STRING FORMAT
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// User authentication
	user := ""
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		// Cookie doesn't exist
	} else {
		user, _ = util.CheckCookie(sessionID.Value, db)
	}

	// HTML generation
	headerInfo := gen.Header {
		Title: "Listings Page",
		User: user,
		Messages: 0,
	}
	cities := gen.ReturnFilter(db)
	listings := gen.ReturnListings(db, query.Origin, query.Destination, query.Time)

	listPage := gen.HeaderHtml(&headerInfo)
	listPage += gen.FilterHtml(cities, query.Origin, query.Destination, util.ReverseConvertDate(query.Time))
	listPage += gen.ListingsHtml(listings)
	listPage += gen.FooterHtml()

	fmt.Fprint(w, listPage)
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request){
	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// User authentication
	user := ""
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		// Cookie doesn't exist
	} else {
		user, _ = util.CheckCookie(sessionID.Value, db)
	}

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}
	// HTML generation
	headerInfo := gen.Header {
		Title: "Create Listing Page",
		User: user,
		Messages: 0,
	}
	createListingsPage := gen.HeaderHtml(&headerInfo)
	cities := gen.ReturnFilter(db)
 	createListingsPage += gen.CreateListingHtml(user, cities)
	createListingsPage += gen.FooterHtml()
	fmt.Fprint(w, createListingsPage)
}

func DashListingsHandler(w http.ResponseWriter, r *http.Request){
	// Database initialization

	fmt.Fprint(w, "Dash Listings Handlerf")
}

func CreateSubmitHandler(w http.ResponseWriter, r *http.Request){
		// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// User authentication
	user := ""
	var userId int
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		// Cookie doesn't exist
	} else {
		user, userId = util.CheckCookie(sessionID.Value, db)
	}

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	// originId, err := strconv.Atoi(r.FormValue("Origin"))
	// if err != nil {
	// 	fmt.Fprint(w, "Invalid origin")
	// 	return
	// }

	// destinationId, err := strconv.Atoi(r.FormValue("Destination"))
	// if err != nil {
	// 	fmt.Fprint(w, "Invalid destination")
	// 	return
	// }
	// seats, err := strconv.Atoi(r.FormValue("Seats"))
	// if err != nil {
	// 	fmt.Fprint(w, "Invalid number of seats")
	// 	return
	// }
	// fee, err := strconv.ParseFloat(r.FormValue("Fee"), 64)
	// if err != nil {
	// 	fmt.Fprint(w, "Invalid fee amount")
	// 	return
	// }


	// if r.FormValue("Leaving") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
	// 	fmt.Fprint(w, "Please fully fill out the form")
	// 	return
	// }
	// if r.FormValue("Origin") == r.FormValue("Destination") {
	// 	fmt.Fprint(w, "Please enter different origins and destinations")
	// 	return
	// }

	err := ValidCreateSubmit(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	dateLeaving := util.ConvertDate(r.FormValue("Leaving"))
	err = util.CompareDate(dateLeaving, time.Now().Local().Format(time.RFC3339))

	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	// if fee > 100 {
	// 	fmt.Fprint(w, "Too high of a fee")
	// 	return
	// }

	// if seats > 8 {
	// 	fmt.Fprint(w, "Too many seats for a legit car")
	// 	return
	// }

	err = gen.CheckNearbyListings(db, r.FormValue("Leaving"), userId)
	if err !=nil {
		fmt.Fprint(w, "You have another listing within an hour of this one.")
		return
	}

	err = gen.CreateListing(db, r.FormValue("Leaving"), userId, originId, destinationId, seats, fee)
	if err!=nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "Created listing!")
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	user := gen.ReturnUserInfo(db, r.URL.Path[3:])
	formatted, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		fmt.Fprint(w, "can't convert to json")
		return
	}
	fmt.Fprint(w, string(formatted))
}

func AppHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Fprint(w, "can't parse url query string")
		return
	}
	query, err := util.ValidListingQuery(u) // Returns util.QueryFields
	if err != nil {
		fmt.Fprint(w, "nonvalid query string")
		return
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		fmt.Fprint(w, "database error 1")
		return
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Fprint(w, "database error 2")
		return
	}

	listings := gen.ReturnListings(db, query.Origin, query.Destination, query.Time)
	jsonListings, err := json.MarshalIndent(listings, "", "    ")
	if err != nil {
		fmt.Fprint(w, "convert to json failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonListings))
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" || r.FormValue("Email") == "" {
		fmt.Fprint(w, "enter a password/username")
		return
	}

	if utf8.RuneCountInString(r.FormValue("Password")) < 6 {
		fmt.Fprint(w, "Password is not long enough")
		return
	}

	if r.FormValue("Password") != r.FormValue("Password2"){
		fmt.Fprint(w, "Passwords did not match")
		return
	}

	if r.FormValue("Email") != r.FormValue("Email2") {
		fmt.Fprint(w, "Emails did not match");
		return
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	err = gen.CheckUserInfo(db, r.FormValue("Username"), r.FormValue("Email"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// Registration details validation
	// if gen.UnusedUsername(db, r.FormValue("Username")){
	// 	fmt.Fprint(w, "Username is taken")
	// 	// http.Redirect(w, r, "https://5sur.com/u=usernameTaken", 301)
	// 	return
	// }

	// if gen.InvalidUsername(r.FormValue("Username")){
	// 	fmt.Fprint(w, "Username is an invalid form")
	// 	return
	// }

	// if gen.UnusedEmail(db, r.FormValue("Email")){
	// 	fmt.Fprint(w, "Email is already registered")
	// 	return
	// }

	// Create user
	gen.UserAuth(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))

	fmt.Fprint(w, "Confirmation email has been sent to " + r.FormValue("Email"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		fmt.Fprint(w, "enter a password/username")
		return
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		fmt.Fprint(w, "SQL ERROR")
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Fprint(w, "Can't establish database connection")
	}

	// User authentication
	if gen.CheckCredentials(db, r.FormValue("Username"), r.FormValue("Password")) {
		myCookie := util.CreateCookie(r.FormValue("Username"), db) // This also stores a hashed cookie in the database
		http.SetCookie(w, &myCookie)
		http.Redirect(w, r, "https://5sur.com/l/", 301)
		return
	}else {
		fmt.Fprint(w, "Your username/password was incorrect")
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Create gen.InvalidateCookie
	authCookie := http.Cookie{
		Name: "RideChile",
		Value: "",
		Path: "/",
		Domain: "5sur.com", // Add domain name in the future
		Expires: time.Now().Add(-1000), // Expire cookie
		MaxAge: -1,
		Secure: true, // SSL only
		HttpOnly: true, // HTTP(S) only
	}
	http.SetCookie(w, &authCookie)

	// CREATE DELETE SESSION FROM SERVER
	fmt.Fprint(w, "you SHOULD be logged out")
}

func AccountAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Query string validation
	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}

	token, err := util.ValidAuthQuery(u) // Returns util.QueryFields
	if err != nil {
		fmt.Fprint(w, "nonvalid query string")
		return
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	user, err := gen.CreateUser(db, token)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, user + ", your accout is activated!")
	return
}

func ReserveFormHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Fprint(w, "Url parse error")
		return
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		fmt.Fprint(w, "Url parse error")
		return
	}
	if _,ok := m["l"]; !ok {
		fmt.Fprint(w, "Missing listing id")
		return
	}
	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// User authentication
	user := ""
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		// Cookie doesn't exist
	} else {
		user, _ = util.CheckCookie(sessionID.Value, db)
	}

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}
	// HTML generation
	headerInfo := gen.Header {
		Title: "Reserve Page",
		User: user,
		Messages: 0,
	}
	reservePage := gen.HeaderHtml(&headerInfo)
	reservePage += gen.ReserveHtml(m["l"][0])
	reservePage += gen.FooterHtml()
	fmt.Fprint(w, reservePage)
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) {
	// Check POST
	if r.FormValue("Seats") == "" || r.FormValue("Listing") == ""{
		fmt.Fprint(w, "Missing required fields")
		return
	}
	
	listingId, err := strconv.Atoi(r.FormValue("Listing"))
	if err != nil {
		fmt.Fprint(w, "Invalid listing")
		return
	}
	
	seats, err := strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		fmt.Fprint(w, "Seat not an integer")
		return
	}

	// Database initialization
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// User authentication
	username := ""
	var userId int
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		// Cookie doesn't exist
	} else {
		username, userId = util.CheckCookie(sessionID.Value, db)
	}

	if username == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	ride, err := gen.ReturnIndividualListing(db, listingId)
	if err != nil {
		fmt.Fprint(w, "listing does not exist")
		return
	}

	if seats > ride.Seats {
		fmt.Fprint(w, "Not enough seats available")
		return
	}
	
	err = gen.ValidReservation(db, userId, listingId, ride.DateLeaving)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	err = gen.MakeReservation(db, listingId, seats, userId, r.FormValue("Message"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// HTML generation
	headerInfo := gen.Header {
		Title: "Reserve Page",
		User: username,
		Messages: 0,
	}

	reservePage := gen.HeaderHtml(&headerInfo)
	// Temp
	reservePage += "<br /><br /><br /><br />Placed on the reservation queue!\r\nListing ID: " + strconv.Itoa(listingId) + "\r\nSeats: " + strconv.Itoa(seats) + "User: " + username + "\r\nMessage: " + r.FormValue("Message")
	reservePage += gen.FooterHtml()

	fmt.Fprint(w, reservePage)
	return
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://5sur.com/l/", 301)
	return
}

func main() {
	http.HandleFunc("/l/", ListingsHandler)
	http.HandleFunc("/u/", UserHandler)
	http.HandleFunc("/a/", AppHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegistrationHandler)
	http.HandleFunc("/auth/", AccountAuthHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/reserveSubmit", ReserveHandler)
	http.HandleFunc("/reserve", ReserveFormHandler)
	http.HandleFunc("/create", CreateListingHandler)
	http.HandleFunc("/createSubmit", CreateSubmitHandler)
	http.HandleFunc("/dash/listings", DashListingsHandler)
	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(":8080", nil)
}