package app

import (
	"fmt"
	"net/http"
	"encoding/json"
	"5sur/gen"
	"5sur/util"
	"strings"
)

func ListingsHandler(w http.ResponseWriter, r *http.Request) error {
	query, err := util.ValidListingQuery(r.URL) // Returns util.QueryFields
	if err != nil { return err }

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	listings, err := gen.ReturnListings(db, query.Origin, query.Destination, query.Date + " " + query.Time)
	if err != nil { return err }
	jsonListings, err := json.MarshalIndent(listings, "", "    ")
	if err != nil {
		return util.NewError(nil, "Json conversion failed", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonListings))
	return nil
}

func CityHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()
	cities, err := gen.ReturnFilter(db)
	if err != nil { return err }

	jsonCities, err := json.MarshalIndent(cities, "", "    ")
	if err != nil {
		return util.NewError(err, "Json conversion failed", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonCities))
	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) error {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		return util.NewError(nil, "Missing username or password", 400)
	}

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()
	
	// User authentication
	authenticated, err := gen.CheckCredentials(db, r.FormValue("Username"), r.FormValue("Password"))
	if err != nil { return err }
	if authenticated {
		myCookie, err := util.CreateCookie(r.FormValue("Username"), db, true, true) // This also stores a hashed cookie in the database
		if err != nil { return err }
		http.SetCookie(w, &myCookie)
		w.WriteHeader(200)
		fmt.Fprint(w, "Logged in as " + r.FormValue("Username"))
		return nil
	} else {
		return util.NewError(nil, "Your username or password was incorrect", 400)
	}
	return nil
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) error {
	//Check POST data
	values, err := util.ValidReservePost(r)
	if err != nil { return err }

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckAppCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	err = gen.CreateReservation(db, userId, values.ListingId, values.Seats, r.FormValue("Message"))
	if err != nil { return err }

	w.WriteHeader(200)
	fmt.Fprint(w, "You registered, woo")
	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err!=nil {
		return err
	}
	defer db.Close()

	// User authentication
	_, userId, err := util.CheckAppCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	// Create gen.InvalidateCookie
	err, expiredCookie := util.DeleteCookie(db, userId, true)
	if err != nil { return err }
	http.SetCookie(w, &expiredCookie)
	w.WriteHeader(200)

	fmt.Fprint(w, "You logged out")
	return nil
}

func UserHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	splits := strings.Split(r.URL.Path, "/")
	user, err := gen.ReturnUserInfo(db, splits[3])
	if err != nil { return err }

	formatted, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return util.NewError(err, "Json conversion failed", 500)
	}

	fmt.Fprint(w, string(formatted))
	return nil
}

func DashListingsHandler(w http.ResponseWriter, r *http.Request) error{
	token, err := util.ValidDashQuery(r.URL)
	specificListing := false
	if err == nil {
		specificListing = true
	} else {
		token = 0
	}
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckAppCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	// Check post data for if a button was clicked that directed the user here.
	if specificListing {
		err = gen.DeleteAlert(db, userId, "dropped", token)
		if err != nil { return err }
		err := gen.CheckPost(db, userId, r, token)
		if err != nil { return err }
	}

	dashListings, err := gen.GetDashListings(db, userId)
	if err != nil { return err }

	var listing gen.SpecificListing
	if specificListing {
		listing, err = gen.SpecificDashListing(db, dashListings, token)
		if err != nil { return err }
	}

	body := &gen.DashListingsHTML{
		SidebarListings: dashListings,
		Listing: listing,
	}

	formatted, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		return util.NewError(err, "Json conversion failed", 500)
	}
	fmt.Fprint(w, string(formatted))
	
	return nil
}