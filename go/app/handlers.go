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