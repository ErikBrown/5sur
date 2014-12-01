package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
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

type DashMessages struct {
	Id int
	Name string
	Picture string
	Count int
	Opened bool
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

func GetDashMessages(db *sql.DB, userId int) ([]DashMessages, error) {
	var results []DashMessages

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT m.sender, u.name, u.picture, count(*), min(m.opened)
			FROM messages as m 
			JOIN users AS u 
				ON u.id = m.sender 
			WHERE m.receiver = ?
			GROUP BY m.sender;
		`)
	if err != nil {
		return results, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return results, err
	}

	for rows.Next() {
		var temp DashMessages
		err := rows.Scan(&temp.Id, &temp.Name, &temp.Picture, &temp.Count, &temp.Opened)
		if err != nil {
			return results, err
		}
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

