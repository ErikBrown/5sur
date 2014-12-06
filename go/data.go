package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/gen"
	"data/util"
	"os"
	// "log"
)

func openDb() (*sql.DB, error) {
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		return db, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}
	return db, nil
}

func AccountAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Query string validation
	token, err := util.ValidAuthQuery(r.URL) // Returns util.QueryFields
	if err != nil {
		fmt.Fprint(w, "nonvalid query string")
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// Authenticate and create the user account
	user, err := gen.CreateUser(db, token)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, user + ", your accout is activated!")
	return
}

func AppHandler(w http.ResponseWriter, r *http.Request) {
	query, err := util.ValidListingQuery(r.URL) // Returns util.QueryFields
	if err != nil {
		fmt.Fprint(w, "nonvalid query string")
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	listings := gen.ReturnListings(db, query.Origin, query.Destination, query.Time)
	jsonListings, err := json.MarshalIndent(listings, "", "    ")
	if err != nil {
		fmt.Fprint(w, "convert to json failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonListings))
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request){
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// User authentication
	user, _ := util.CheckCookie(r, db) // return "" if not logged in
	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	// HTML generation (also does listing-specific SQL calls)
	createListingPage, err := gen.CreateListingPage(db, user);
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, createListingPage)
}

func CreateSubmitHandler(w http.ResponseWriter, r *http.Request){
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	// User authentication
	user, userId := util.CheckCookie(r, db) // return "" if not logged in
	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	createFormPost, err := util.ValidCreateSubmit(r)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	err = gen.CreateListing(db, createFormPost.Date, userId, createFormPost.Origin, createFormPost.Destination, createFormPost.Seats, createFormPost.Fee)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, "listing created!")
}

func DashListingsHandler(w http.ResponseWriter, r *http.Request){
	token, err := util.ValidDashQuery(r.URL)
	specificListing := false
	if err == nil {
		specificListing = true
	} else {
		token = 0
	}
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	// User authentication
	user, userId := util.CheckCookie(r, db) // return "" if not logged in

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	// Check post data for if a button was clicked that directed the user here.
	if specificListing {
		err := gen.CheckPost(db, userId, r, token)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
	}

	dashListings, err := gen.GetDashListings(db, userId)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	var listing gen.SpecificListing
	if specificListing {
		// DO SPECIFIC LISTING
		listing, err = gen.SpecificDashListing(db, dashListings, token)
	}

	// HTML generation
	dashListingsPage, err := gen.DashListingsPage(dashListings, listing, user);
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, dashListingsPage)
}

func DashMessagesHandler(w http.ResponseWriter, r *http.Request){
	// Database initialization
	db, err := openDb()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	defer db.Close()

	// User authentication
	user, userId := util.CheckCookie(r, db) // return "" if not logged in

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}
	dashMessages, err := gen.GetDashMessages(db, userId)
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

	formatted, err := json.MarshalIndent(dashMessages, "", "    ")
	if err != nil {
		fmt.Fprint(w, "can't convert to json")
		return
	}
	fmt.Fprint(w, string(formatted))
}

func ListingsHandler(w http.ResponseWriter, r *http.Request) {
	// log.Println("sdfsdf")
	
	// Convert POST to GET (also does a time validation)
	if r.FormValue("Origin") != "" && r.FormValue("Destination") != "" {
		convertedDate := ""
		convertedTime := ""
		var err error
		if r.FormValue("Date") == "" {
			convertedDate, convertedTime = util.ReturnCurrentTimeString()
		} else if r.FormValue("Time") == "" {
			convertedDate, _, err = util.ReturnTimeString(false, r.FormValue("Date"), "00:00")
			currentDate, currentTime := util.ReturnCurrentTimeString()
			if currentDate == convertedDate {
				convertedTime = currentTime
			} else {
				convertedTime = "00:00"
			}
		} else {
			convertedDate, convertedTime, err = util.ReturnTimeString(false, r.FormValue("Date"), r.FormValue("Time"))
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		}
		http.Redirect(w, r, "https://5sur.com/l/?o=" + r.FormValue("Origin") + "&d=" + r.FormValue("Destination") + "&t=" + convertedDate + "&h=" + convertedTime, 301)
		return
	}

	// Query string validation
	query, err := util.ValidListingQuery(r.URL)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// User authentication
	user, _ := util.CheckCookie(r, db) // return "" if not logged in

	// HTML generation (also does listing-specific SQL calls)
	listPage, err := gen.ListingsPage(db, query, user);
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, listPage)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		fmt.Fprint(w, "enter a password/username")
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

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
	expiredCookie := util.DeleteCookie()
	http.SetCookie(w, &expiredCookie)

	fmt.Fprint(w, "you SHOULD be logged out")
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	err := util.ValidRegister(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	err = gen.CheckUserInfo(db, r.FormValue("Username"), r.FormValue("Email"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	
	gen.UserAuth(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))

	fmt.Fprint(w, "Confirmation email has been sent to " + r.FormValue("Email"))
}

func ReserveFormHandler(w http.ResponseWriter, r *http.Request) {
	l, err := util.ValidReserveURL(r)
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// User authentication
	user, _ := util.CheckCookie(r, db) // return "" if not logged in

	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	reservePage := gen.CreateReserveFormPage(l, user)
	fmt.Fprint(w, reservePage)
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) {
	//Check POST data
	values, err := util.ValidRegisterPost(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// User authentication
	user, userId := util.CheckCookie(r, db) // return "" if not logged in
	if user == "" {
		fmt.Fprint(w, "not logged in")
		return
	}

	err = gen.CreateReservation(db, userId, values.ListingId, values.Seats, r.FormValue("Message"))
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	reservePage := gen.CreateReservePage(values.ListingId, values.Seats, user, r.FormValue("Message"))

	fmt.Fprint(w, reservePage)
	return
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	// User authentication
	user, _ := util.CheckCookie(r, db) // return "" if not logged in

	// HTML generation (also does listing-specific SQL calls)
	homePage, err := gen.HomePage(db, user);
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	fmt.Fprint(w, homePage)
	return
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	// Database initialization
	db, err := openDb()
	if err!=nil {
		fmt.Fprint(w, err)
		return
	}
	defer db.Close()

	user := gen.ReturnUserInfo(db, r.URL.Path[3:])
	formatted, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		fmt.Fprint(w, "can't convert to json")
		return
	}
	fmt.Fprint(w, string(formatted))
}

func EnvHandler(w http.ResponseWriter, r *http.Request) {
	// Database initialization
	
	fmt.Fprint(w, "hi" + os.Getenv("TEST"))
}

func main() {
	util.ConfigureLog()
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
	http.HandleFunc("/dashboard/listings", DashListingsHandler)
	http.HandleFunc("/dashboard/messages", DashMessagesHandler)
	http.HandleFunc("/env", EnvHandler)
	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(":8080", nil)
}