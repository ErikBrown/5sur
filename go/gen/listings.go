package gen

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/util"
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

func ReturnListings(o int, d int) string {
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
	resultString := ""
	for i := range results{
		resultString += GenerateListing(results[i])
	}
	return resultString
}

func GenerateListing (myListing Listing) string{
	var date util.Date = util.CustomDate(myListing.DateLeaving)
	output := `
	<ul class="list_item">
		<li class="listing_user">
			<img src="https://192.241.219.35/` + myListing.Picture + `" alt="User Picture">
			<span class="positive">+100</span>
		</li>
		<li class="date_leaving">
			<div>
				<span class="month">` + date.Month + `</span>
				<span class="day">` + date.Day + `</span>
				<span class="time">` + date.Time + `</span>
			</div>
		</li>
		<li class="city">
			<span>` + myListing.Origin + `</span>
			<span class="to">&#10132;</span>
			<span>` + myListing.Destination + `</span>
		</li>
		<li class="seats">
			<span>` + fmt.Sprintf("%d", myListing.Seats) + `</span>
		</li>
			<li class="fee"><span>$` + fmt.Sprintf("%.2f", myListing.Fee) + `</span>
		</li>
	</ul>
	`
	return output
}