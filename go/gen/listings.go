package gen

import (
	"database/sql"
	"data/util"
	_ "github.com/go-sql-driver/mysql"
)

func ReturnFilter(db *sql.DB) ([]City, error) {
	results := make ([]City, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT * from cities
			ORDER BY name
	`)

	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query()
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp City
		err := rows.Scan(&temp.Id, &temp.Name)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		results = append(results, temp)
	}
	return results, nil
} 

func checkNearbyListings(db *sql.DB, date_leaving string, id int) error {
	stmt, err := db.Prepare(`
		SELECT * FROM listings 
		WHERE date_leaving <= ? + INTERVAL 1 HOUR AND 
		date_leaving >= ? - INTERVAL 1 HOUR AND
		driver = ?
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(date_leaving, date_leaving, id)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		return util.NewError(err, "You already have a listing near this date", 400)
	}
	return nil
}

func ReturnListings(db *sql.DB, o int, d int, t string) ([]Listing, error) {
	results := make ([]Listing, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.picture, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.origin = ? 
				AND l.destination = ? 
				AND l.date_leaving >= ?
				AND l.seats > 0
				AND l.date_leaving > NOW()
			ORDER BY l.date_leaving
			LIMIT 25
		`)
	
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(o, d, t)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp Listing
		err := rows.Scan(&temp.Id, &temp.Driver, &temp.Picture, &temp.Timestamp, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		prettyTime, err := util.PrettyDate(temp.Timestamp, false)
		if err != nil { return results, err }
		temp.Date = prettyTime.Month + " " + prettyTime.Day
		temp.Time = prettyTime.Time
		results = append(results, temp)
	}
	return results, nil
}

func CreateListing(db *sql.DB, date_leaving string, driver int, origin int, destination int, seats int, fee float64) error {
	// This needs to take in account the hour/minute!!! Concatinate the form values! CHANGE THIS
	err := checkNearbyListings(db, date_leaving, driver)
	if err !=nil {
		return err // err is already util.MyError
	}

	stmt, err := db.Prepare(`
		INSERT INTO listings (date_leaving,driver,origin,destination,seats,fee,reserved)
			VALUES (?, ?, ?, ?, ?, ?, false)
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(date_leaving, driver, origin, destination, seats, fee)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	
	return nil
}

func ReturnIndividualListing(db *sql.DB, id int) (Listing, error) {
	result := Listing{}
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.picture, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.id = ?
		`)
	
	if err != nil {
		return Listing{}, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&result.Id, &result.Driver, &result.Picture, &result.Timestamp, &result.Origin, &result.Destination, &result.Seats, &result.Fee)
	prettyTime, err := util.PrettyDate(result.Timestamp, false)
	if err != nil { return result, err }
	result.Date = prettyTime.Month + " " + prettyTime.Day
	result.Time = prettyTime.Time

	if err != nil {
		return Listing{}, util.NewError(nil, "Listing does not exist", 400)
	}
	return result, nil
}