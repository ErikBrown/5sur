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
	query, err := util.ValidQueryString(u) // Returns util.QueryFields
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
		user = util.CheckCookie(sessionID.Value, db)
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
	listPage += gen.FilterHTML(cities, query.Origin, query.Destination, util.ReverseConvertDate(query.Time))
	listPage += gen.ListingsHTML(listings)
	listPage += gen.FooterHtml()

	fmt.Fprint(w, listPage)
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
	query, err := util.ValidQueryString(u) // Returns util.QueryFields
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
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		fmt.Fprint(w, "enter a password/username")
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

	// Registration details validation
	if gen.UnusedUsername(db, r.FormValue("Username")){
		fmt.Fprint(w, "Username is taken")
		// http.Redirect(w, r, "https://5sur.com/u=usernameTaken", 301)
		return
	}

	// Create user
	gen.CreateUser(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))
	
	http.Redirect(w, r, "https://5sur.com/l/", 301)
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
	http.Redirect(w, r, "https://5sur.com/l/", 301)
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is the reserve handler")
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
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/reserve", ReserveHandler)
	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(":8080", nil)
}