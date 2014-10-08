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
	// Checks if the query string follows the correct format and converts query variables to integers
	postFields, err := util.ValidPost(r.FormValue("Origin"), r.FormValue("Destination"))
	if err != nil {
		fmt.Fprint(w, gen.Error404())
		return
	}
	if postFields.Origin != 0 {
		http.Redirect(w, r, "https://192.241.219.35/go/l/?o=" + fmt.Sprintf("%d", postFields.Origin) + "&d=" + fmt.Sprintf("%d", postFields.Destination), 301)
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

/*
func ListingsHandler(w http.ResponseWriter, r *http.Request) {
	// Generate user page
}
*/

/*
func ListingsHandler(w http.ResponseWriter, r *http.Request) {
	// Create session cookie and redirect to Listings homepage
}
*/

func main() {
	http.HandleFunc("/go/l/", ListingsHandler)
	http.ListenAndServe(":8080", nil)
}