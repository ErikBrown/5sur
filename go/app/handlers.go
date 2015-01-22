package app

import (
	"fmt"
	"net/http"
	"encoding/json"
	"data/gen"
	"data/util"
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
		myCookie, err := util.CreateCookie(r.FormValue("Username"), db, true) // This also stores a hashed cookie in the database
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
	values, err := util.ValidRegisterPost(r)
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