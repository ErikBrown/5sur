package results

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Listing struct {
	Driver int
	Picture string
	DateLeaving string
	Origin string
	Destination string
	Seats int
	Fee float32
}

func ReturnListings(o string, d string) []Listing {
	results := make ([]Listing, 0)
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

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT u.id, u.picture, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.origin = ? AND l.destination = ? AND DATE(l.date_leaving) >= '2012-10-01 12:30:00'
			ORDER BY l.date_leaving
			LIMIT 25;
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(o, d)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp Listing
		err := rows.Scan(&temp.Driver, &temp.Picture, &temp.DateLeaving, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results
}