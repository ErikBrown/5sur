package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/util"
)

func CreateReservation(db *sql.DB, userId int, listingId int, seats int, message string) error {
	ride, err := ReturnIndividualListing(db, listingId)
	if err != nil {
		return err
	}

	if userId == ride.Driver {
		return util.NewError(nil, "Cannot register for a ride you own", 400)
	}

	if seats > ride.Seats {
		return util.NewError(nil, "Not enough seats available", 400)
	}

	if seats <= 0 {
		return util.NewError(nil, "You must register for at least one seat", 400)
	}
	
	err = validReservation(db, userId, listingId, ride.DateLeaving)
	if err != nil {
		return err
	}

	err = makeReservation(db, listingId, seats, userId, message)
	if err != nil {
		return err
	}
	return nil
}

func validReservation(db *sql.DB, userId int, listingId int, date string) error {
	stmt, err := db.Prepare(`
		SELECT r.id
			FROM reservation_queue as r
			WHERE r.listing_id = ? AND r.passenger_id = ?
	`)
	
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	t := ""
	err = stmt.QueryRow(listingId, userId).Scan(&t)
	if err == nil {
		return util.NewError(nil, "You are already on this reservation queue", 400)
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
		return util.NewError(err, "Database error", 500)
	}
	defer stmt2.Close()

	rows, err := stmt2.Query(userId, date, date)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	results := make ([]int, 0)
	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp int
		err := rows.Scan(&temp)
		if err != nil {
			return util.NewError(err, "Database error", 500)
		}
		results = append(results, temp)
	}
	for _, v := range results {
		if v == listingId {
			return util.NewError(nil, "You are already registered for this listing", 400)
		}
	}

	if len(results) != 0 {
		return util.NewError(nil, "You are already registered for a ride at this time", 400)
	}

	return nil
}

func CheckReservationQueue(db *sql.DB, listingId int) (bool, error) {
	stmt, err := db.Prepare(`
		SELECT * FROM reservation_queue 
			WHERE listing_id = ?
		`)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()
	rows, err := stmt.Query(listingId)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}

	defer rows.Close()

	for rows.Next() {
		return true, nil
	}
	return false, nil
}

func makeReservation(db *sql.DB, listingId int, seats int, userId int, message string) error{
	stmt, err := db.Prepare(`
		INSERT INTO reservation_queue (listing_id, seats, passenger_id, message)
			VALUES (?, ?, ?, ?)
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(listingId, seats, userId, message)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}