package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
)

func ReturnFilter(db *sql.DB) []City {
	results := make ([]City, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT * from cities
			ORDER BY name
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
		var temp City
		err := rows.Scan(&temp.Id, &temp.Name)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results
} 

func checkNearbyListings(db *sql.DB, date_leaving string, id int) error {
	stmt, err := db.Prepare(`
		SELECT * FROM listings 
		WHERE date_leaving <= ? + INTERVAL 1 HOUR AND 
		date_leaving >= ? - INTERVAL 1 HOUR AND
		driver = ?
	`)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(date_leaving, date_leaving, id)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	for rows.Next() {
		return errors.New("You have a listing near this date.")
	}
	return nil
}

func ReturnListings(db *sql.DB, o int, d int, t string) []Listing {
	results := make ([]Listing, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.picture, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.origin = ? AND l.destination = ? AND l.date_leaving >= ?
			ORDER BY l.date_leaving
			LIMIT 25
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(o, d, t)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp Listing
		err := rows.Scan(&temp.Id, &temp.Driver, &temp.Picture, &temp.DateLeaving, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results
}

func CreateListing(db *sql.DB, date_leaving string, driver int, origin int, destination int, seats int, fee float64) error {
	// This needs to take in account the hour/minute!!! Concatinate the form values! CHANGE THIS
	err := checkNearbyListings(db, date_leaving, driver)
	if err !=nil {
		return err
	}

	stmt, err := db.Prepare(`
		INSERT INTO listings (date_leaving,driver,origin,destination,seats,fee,reserved)
			VALUES (?, ?, ?, ?, ?, ?, false)
		`)
	defer stmt.Close()

	if err != nil {
		return err
	}
	_, err = stmt.Exec(date_leaving, driver, origin, destination, seats, fee)
	if err != nil {
		return err	
	}

	return nil
}

func ReturnIndividualListing(db *sql.DB, id int) (Listing, error) {
	result :=Listing{}
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.picture, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.id = ?
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	err = stmt.QueryRow(id).Scan(&result.Id, &result.Driver, &result.Picture, &result.DateLeaving, &result.Origin, &result.Destination, &result.Seats, &result.Fee)
	if err != nil {
		return result, err
	}
	return result, nil
}