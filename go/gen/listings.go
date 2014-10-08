package gen

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/util"
)

type city struct {
	id int
	name string
}

type listing struct {
	driver int
	picture string
	dateLeaving string
	origin string
	destination string
	seats int
	fee float32
}

func ReturnFilter(db *sql.DB, o int, d int) string {
	results := make ([]city, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
SELECT * from cities
	ORDER BY name;
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp city
		err := rows.Scan(&temp.id, &temp.name)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	resultString := `
	<div id="search_wrapper">
		<form method="post" action="https://192.241.219.35/go/l/">
			<select name="Origin">`
	for i := range results{
		resultString += generateOption(results[i], o)
	}
	resultString += `
			</select>
			To
			<select name="Destination">
	`
	for i := range results{
		resultString += generateOption(results[i], d)
	}
	resultString += `
			</select>
			<input type="submit" value="Go">
		</form>
	</div>`
	return resultString
} 

func ReturnListings(db *sql.DB, o int, d int) string {
	results := make ([]listing, 0)

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
		var temp listing
		err := rows.Scan(&temp.driver, &temp.picture, &temp.dateLeaving, &temp.origin, &temp.destination, &temp.seats, &temp.fee)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	resultString := ""
	for i := range results{
		resultString += generateListing(results[i])
	}
	return resultString
}

func generateOption(f city, i int) string {
	selected := ""
	if f.id == i {
		selected = " selected"
	}
	return `<option value=` + fmt.Sprintf("%d", f.id) + selected + `>` + f.name + `</option>`
}

func generateListing (myListing listing) string{
	var date util.Date = util.CustomDate(myListing.dateLeaving)
	output := `
	<ul class="list_item">
		<li class="listing_user">
			<img src="https://192.241.219.35/` + myListing.picture + `" alt="User Picture">
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
			<span>` + myListing.origin + `</span>
			<span class="to">&#10132;</span>
			<span>` + myListing.destination + `</span>
		</li>
		<li class="seats">
			<span>` + fmt.Sprintf("%d", myListing.seats) + `</span>
		</li>
			<li class="fee"><span>$` + fmt.Sprintf("%.2f", myListing.fee) + `</span>
		</li>
	</ul>
	`
	return output
}