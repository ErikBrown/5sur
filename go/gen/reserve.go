package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
)

func ValidReservation(db *sql.DB, userId int, listingId int, date string) error {
	stmt, err := db.Prepare(`
		SELECT r.id
			FROM reservation_queue as r
			WHERE r.listing_id = ? AND r.user_id = ?
	`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	t := ""
	err = stmt.QueryRow(listingId, userId).Scan(&t)
	if err == nil {
		e := errors.New("You are already on this reservation queue")
		return e
	}

	stmt2, err := db.Prepare(`
		SELECT r.listing_id
			FROM reservations as r
			JOIN listings as l ON r.listing_id = l.id
			WHERE 
				r.passenger_id = ? AND 
				l.date_leaving <= ? + INTERVAL 1 HOUR AND 
				l.date_leaving >= ? - INTERVAL 1 HOUR
	`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt2.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt2.Query(userId, date, date)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	results := make ([]int, 0)
	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp int
		err := rows.Scan(&temp)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	for _, v := range results {
		if v == listingId {
			e := errors.New("You are already registered for this listing")
			return e
		}
	}

	if len(results) != 0 {
		e := errors.New("You are already registered for a ride at this time")
		return e
	}

	return nil
}

func MakeReservation(db *sql.DB, listingId int, seats int, userId int, message string) error{
	stmt, err := db.Prepare(`
		INSERT INTO reservation_queue (listing_id, seats, user_id, message)
			VALUES (?, ?, ?, ?)
		`)
	defer stmt.Close()

	if err != nil {
		return err
	}
	_, err = stmt.Exec(listingId, seats, userId, message)
	if err != nil {
		return err
	}
	return nil
}