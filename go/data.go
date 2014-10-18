package main

import (
	"fmt"
	"net/http"
	"net/url"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/gen"
	"data/util"
)

func ListingsHandler(w http.ResponseWriter, r *http.Request) {
	
	// POST validation
	if r.FormValue("Origin") != "" || r.FormValue("Destination") != "" {
		http.Redirect(w, r, "https://192.241.219.35/go/l/?o=" + r.FormValue("Origin") + "&d=" + r.FormValue("Destination"), 301)
		return
	}


	// Query string validation
	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}
	query, err := util.ValidQueryString(u) // Returns util.QueryFields
	if err != nil {
		fmt.Fprint(w, gen.Error404())
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
		user = util.CheckCookie(sessionID.Value, db)
	}

	// HTML generation
	headerInfo := gen.Header {
		Title: "Listings Page",
		User: user,
		Messages: 0,
	}
	cities := gen.ReturnFilter(db, query.Origin, query.Destination)
	listings := gen.ReturnListings(db, query.Origin, query.Destination)

	listPage := gen.HeaderHtml(&headerInfo)
	listPage += gen.FilterHTML(cities, query.Origin, query.Destination)
	listPage += gen.ListingsHTML(listings)
	listPage += gen.FooterHtml()

	fmt.Fprint(w, listPage)
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
		// http.Redirect(w, r, "https://192.241.219.35/u=usernameTaken", 301)
		return
	}

	// Create user
	gen.CreateUser(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))
	
	fmt.Fprint(w, "Success!")
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprint(w, "You're logged in!")
		return
	}else {
		fmt.Fprint(w, "Your username/password was incorrect")
		return
	}
}

func main() {
	http.HandleFunc("/go/l/", ListingsHandler)
	http.HandleFunc("/go/u/", UsersHandler)
	http.HandleFunc("/go/r/", RegistrationHandler)
	http.ListenAndServe(":8080", nil)
}