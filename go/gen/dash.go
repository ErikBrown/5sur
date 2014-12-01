package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
	"net/http"
	"strconv"
)

type DashListing struct {
	Day string
	Month string
	Origin int
	Destination int
	Alert bool
	ListingId int
	Seats int
	Fee int
	Time string
}

type SpecificListing struct {
	Day string
	Month string
	Origin int
	Destination int
	Alert bool
	ListingId int
	Seats int
	Fee int
	Time string
	PendingUsers []PendingUser
	RegisteredUsers []RegisteredUser
}

type PendingUser struct {
		Id int
		Name string
		Picture string
		Message string
}

type RegisteredUser struct{
		Id int
		Name string
		Picture string
}

func GetDashListings(db *sql.DB, userId int) ([]DashListing, error) {
	results := make ([]DashListing, 0)

		// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.date_leaving, l.origin, l.destination, l.id, l.seats, l.fee
			FROM listings AS l
			WHERE l.date_leaving >= NOW() AND l.driver = ?
			ORDER BY l.date_leaving
			LIMIT 25
		`)

	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

		// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		date := ""
		var temp DashListing
		err := rows.Scan(&date, &temp.Origin, &temp.Destination, &temp.ListingId, &temp.Seats, &temp.Fee)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		convertedDate := util.PrettyDate(date, false)
		temp.Day = convertedDate.Day
		temp.Month = convertedDate.Month
		temp.Time = convertedDate.Time
		temp.Alert, err = CheckReservationQueue(db, temp.ListingId)
		if err != nil {
			return results, err
		}
		// Also find if there are any new messages.
		results = append(results, temp)
	}

	return results, nil

}

func SpecificDashListing(db *sql.DB, listings []DashListing, listingId int) (SpecificListing, error) {
	found := false
	var err error
	var myListing DashListing
	for i := range listings{
		if listings[i].ListingId == listingId {
			myListing = listings[i]
			found = true
			break
		}
	}
	if !found {
		return SpecificListing{}, errors.New("Could not find specific listing")
	}
	var result SpecificListing
	result.Day = myListing.Day
	result.Month = myListing.Month
	result.Origin = myListing.Origin
	result.Time = myListing.Time
	result.Alert = myListing.Alert
	result.ListingId = myListing.ListingId
	result.Seats = myListing.Seats
	result.Fee = myListing.Fee
	result.Destination = myListing.Destination

	result.PendingUsers, err = getPendingUsers(db, listingId)
	if err != nil {
		return result, err
	}
	result.RegisteredUsers, err = getRegisteredUsers(db, listingId)
	if err != nil {
		return result, err
	}
	return result, nil

}

func getPendingUsers(db *sql.DB, listingId int) ([]PendingUser, error) {
	stmt, err := db.Prepare(`
		SELECT r.message, u.id, u.name, u.picture
			FROM reservation_queue as r
			JOIN users AS u ON r.passenger_id = u.id
			WHERE r.listing_id = ?;
	`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()
	var results []PendingUser
	for rows.Next() {
		var temp PendingUser
		err := rows.Scan(&temp.Message, &temp.Id, &temp.Name, &temp.Picture)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results, nil
}

func getRegisteredUsers(db *sql.DB, listingId int) ([]RegisteredUser, error) {
	stmt, err := db.Prepare(`
		SELECT u.id, u.name, u.picture
			FROM reservations as r
			JOIN users AS u ON r.passenger_id = u.id
			WHERE r.listing_id = ?;
	`)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer rows.Close()

	var results []RegisteredUser
	for rows.Next() {
		var temp RegisteredUser
		err := rows.Scan( &temp.Id, &temp.Name, &temp.Picture)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results, nil
}

func deleteFromQueue(db *sql.DB, userId int, listingId int, passenger_id int) (bool, error) {
	stmt, err := db.Prepare(`
		DELETE FROM reservation_queue 
			WHERE passenger_id IN 
				(SELECT * FROM
					(SELECT r.passenger_id 
					FROM reservation_queue AS r
					JOIN listings as l 
							ON l.id = r.listing_id 
						JOIN users as u 
							ON l.driver = u.id 
						WHERE r.passenger_id = ? and u.id = ?
						AND l.id = ?)
				AS s) 
			AND listing_id = ?;
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	affected, err := stmt.Exec(passenger_id, userId, listingId, listingId)
	if err != nil {
		return false, err
	}
	rowsDeleted, err := affected.RowsAffected()
	if err != nil {
		panic(err.Error())
	}
	if rowsDeleted == 0{
		return false, nil
	}
	return true, nil
}

func addToReservation(db *sql.DB, userId int, listingId int, passenger_id int) error {
		stmt, err := db.Prepare(`
		INSERT INTO reservations(listing_id, driver_id, passenger_id)
			SELECT * FROM (SELECT ?, ?, ?) AS tmp
			WHERE NOT EXISTS (
				SELECT listing_id
					FROM reservations
					WHERE listing_id = ?
					AND driver_id = ?
					AND passenger_id = ?
				) LIMIT 1;
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(listingId, userId, passenger_id, listingId, userId, passenger_id)
	if err != nil {
		return err
	}
	return nil
}

func CheckPost(db *sql.DB, userId int, r *http.Request, listingId int) error {
	if r.FormValue("a") != "" {
		add, err := strconv.Atoi(r.FormValue("a"))
		if err != nil {
			return errors.New("Invalid")
		}
		deleted, err := deleteFromQueue(db, userId, listingId, add)
		if err != nil {
			return err
		}
		if deleted {
			err := addToReservation(db, userId, listingId, add)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if r.FormValue("r") != "" {
		remove, err := strconv.Atoi(r.FormValue("r"))
		if err != nil {
			return errors.New("Invalid")
		}	
		_, err = deleteFromQueue(db, userId, listingId, remove)
		if err != nil {
			return err
		}
		return nil
	}
	if r.FormValue("m") != "" {
		_, err := strconv.Atoi(r.FormValue("m"))
		if err != nil {
			return errors.New("Invalid")
		}
		// Deal with messenging
	}
	return nil
}

