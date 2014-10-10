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
	if r.FormValue("Origin") != "" {
		http.Redirect(w, r, "https://192.241.219.35/go/l/?o=" + r.FormValue("Origin") + "&d=" + r.FormValue("Destination"), 301)
		return
	}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		// panic
	}
	query, err := util.ValidQueryString(u)
	if err != nil {
		fmt.Fprint(w, gen.Error404())
		return
	}
	// The db should be long lived. Do not recreate it unless accessing a different
	// database. Do not Open() and Close() from a short lived function, just pass in
	// the db object to the function
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// Defer Close() to be run at the end of main()
	defer db.Close()

	// sql.Open does not establish any connections to the database - To check if the
	// database is available and accessable, use sql.Ping()
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// Generate the html for the listings page
	myString := gen.HeaderHtml("Listings Page")
	myString += gen.ReturnFilter(db, query.Origin, query.Destination)
	myString += gen.ReturnListings(db, query.Origin, query.Destination)
	myString += gen.FooterHtml()

	// Print html
	fmt.Fprint(w, myString)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
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


	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		fmt.Fprint(w, "enter a password/username")
		return
	}

	if gen.UnusedUsername(db, r.FormValue("Username")){
		fmt.Fprint(w, "Username is taken")
		// http.Redirect(w, r, "https://192.241.219.35/u=usernameTaken", 301)
		return
	}


	gen.CreateUser(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))
	fmt.Fprint(w, "Success!")
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		fmt.Fprint(w, "enter a password/username")
		return
	}

	if gen.CheckCredentials(db, r.FormValue("Username"), r.FormValue("Password")) {
		fmt.Fprint(w, "You're a user!")
	}else {
		fmt.Fprint(w, "Your username/password was incorrect")
	}

}

func main() {
	http.HandleFunc("/go/l/", ListingsHandler)
	http.HandleFunc("/go/u/", UsersHandler)
	http.HandleFunc("/go/r/", RegistrationHandler)
	http.ListenAndServe(":8080", nil)
}