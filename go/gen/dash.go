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
	Origin string
	Destination string
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

type DashReservation struct {
	ListingId int
	Time string
	Origin string
	Destination string
	Seats string
	Fee string
}

type PendingUser struct {
	Id int
	Name string
	Picture string
	Message string
	Seats int
}

type RegisteredUser struct{
	Id int
	Name string
	Picture string
	Seats int
}

type Reservation struct {
	Time string
	Origin string
	Destination string
	Seats string
	Fee string
	ListingId int
	DriverId int
	DriverName string
	DriverPicture string
}

type SpecificListing struct {
	Day string
	Month string
	Origin string
	Destination string
	Alert bool
	ListingId int
	Seats int
	Fee int
	Time string
	PendingUsers []PendingUser
	RegisteredUsers []RegisteredUser
}

type SpecificMessage struct {
	Id int
	Name string
	Picture string
	Date string
	Message string
}

func GetDashListings(db *sql.DB, userId int) ([]DashListing, error) {
	results := make ([]DashListing, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.date_leaving, c.name, c2.name, l.id, l.seats, l.fee
			FROM listings AS l
			JOIN cities as c ON l.origin = c.id
			LEFT JOIN cities as c2 ON l.destination = c2.id
			WHERE l.date_leaving >= NOW() AND l.driver = ?
			ORDER BY l.date_leaving;
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

func GetDashReservations(db *sql.DB, userId int) ([]DashReservation, error) {
	results := make ([]DashReservation, 0)
	stmt, err := db.Prepare(`
		SELECT l.id, l.date_leaving, l.origin, l.destination, l.seats, l.fee
			FROM listings AS l
			JOIN reservations as r ON l.id = r.listing_id
			WHERE r.passenger_id = ?
			ORDER BY l.date_leaving;
		`)

	if err != nil {
		return results, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return results, err
	}

		// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		temp := DashReservation{}
		err := rows.Scan(&temp.ListingId, &temp.Time, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, err
		}
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

func SpecificDashMessage(db *sql.DB, messages []DashMessages, messageId int) (SpecificMessage, error) {
	found := false
	message := SpecificMessage{}
	for i := range messages{
		if messages[i].Id == messageId {
			message.Id = messages[i].Id
			message.Name = messages[i].Name
			message.Picture = messages[i].Picture
			found = true
			break
		}
	}
	if !found {
		return SpecificMessage{}, errors.New("Could not find specific message")
	}

	stmt, err := db.Prepare(`
		SELECT m.message
			FROM messages as m 
			WHERE m.id = ?;
		`)
	if err != nil {
		return SpecificMessage{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(messageId)
	if err != nil {
		return SpecificMessage{}, err
	}

	for rows.Next() {
		err := rows.Scan(&message.Message)
		if err != nil {
			return SpecificMessage{}, err
		}
	}

	return message, nil
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

func SpecificDashReservation(db *sql.DB, reservations []DashReservation, listingId int) (Reservation, error) {
	found := false
	results := Reservation{}
	for i := range reservations{
		if reservations[i].ListingId == listingId {
			results.Time = reservations[i].Time
			results.Origin = reservations[i].Origin
			results.Destination = reservations[i].Destination
			results.Seats = reservations[i].Seats
			results.Fee = reservations[i].Fee
			results.ListingId = reservations[i].ListingId
			found = true
			break
		}
	}

	if !found {
		return Reservation{}, errors.New("Could not find specific reservation")
	}

	stmt, err := db.Prepare(`
		SELECT u.id, u.name, u.picture
			FROM users as u
			LEFT JOIN reservations as r ON u.id = r.driver_id
			WHERE r.listing_id = ?
		`)
	if err != nil {
		return Reservation{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		return Reservation{}, err
	}

	for rows.Next() {
		err := rows.Scan(&results.DriverId, &results.DriverName, &results.DriverPicture)
		if err != nil {
			return Reservation{}, err
		}
	}

	return results, nil
}

func getPendingUsers(db *sql.DB, listingId int) ([]PendingUser, error) {
	stmt, err := db.Prepare(`
		SELECT r.message, u.id, u.name, u.picture, r.seats
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
		err := rows.Scan(&temp.Message, &temp.Id, &temp.Name, &temp.Picture, &temp.Seats)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}
	return results, nil
}

func getPendingUser(db *sql.DB, listingId int, pendingUserId int) (PendingUser, error) {
	stmt, err := db.Prepare(`
		SELECT r.message, u.id, u.name, u.picture, r.seats
			FROM reservation_queue as r
			JOIN users AS u ON r.passenger_id = u.id
			WHERE r.listing_id = ? AND u.id = ?;
	`)
	
	if err != nil {
		return PendingUser{}, err
	}
	defer stmt.Close()

	pendingUser := PendingUser{}
	err = stmt.QueryRow(listingId, pendingUserId).Scan(&pendingUser.Message, &pendingUser.Id, &pendingUser.Name, &pendingUser.Picture, &pendingUser.Seats)
	if err != nil {
		return pendingUser, err
	}
	return pendingUser, nil
}

func getRegisteredUsers(db *sql.DB, listingId int) ([]RegisteredUser, error) {
	stmt, err := db.Prepare(`
		SELECT u.id, u.name, u.picture, r.seats
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
		err := rows.Scan( &temp.Id, &temp.Name, &temp.Picture, &temp.Seats)
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

func deleteFromReservations(db *sql.DB, userId int, listingId int, passengerId int) (bool, error) {
	stmt, err := db.Prepare(`
		DELETE FROM reservations
			WHERE driver_id = ?
				AND listing_id = ?
				AND passenger_id = ?
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	affected, err := stmt.Exec(userId, listingId, passengerId)
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

func addToReservation(db *sql.DB, userId int, listingId int, passengerId int, seats int) error {
		stmt, err := db.Prepare(`
		INSERT INTO reservations(listing_id, driver_id, passenger_id, seats)
			SELECT ? AS listing_id, ? AS driver_id, ? AS passenger_id, ? AS seats FROM dual
				 HERE NOT EXISTS (
					SELECT listing_id
						FROM reservations
						WHERE listing_id = ?
						AND driver_id = ?
						AND passenger_id = ?
				) LIMIT 1;
		`)
	
	if err != nil {
		return err // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(listingId, userId, passengerId, seats, listingId, userId, passengerId)
	if err != nil {
		return err
	}
	return nil
}

func updateSeats(db *sql.DB, userId int, listingId int, seats int) error {
	stmt, err := db.Prepare(`
		UPDATE listings
			SET seats = seats + ?
			WHERE id = ?
				AND driver = ?;
			`)
	if err != nil {
		return err // Have a proper error in production
	}
	defer stmt.Close()
	_, err = stmt.Exec(seats, listingId, userId)
	if err != nil {
		return err
	}
	return nil
}

func findSeats(db *sql.DB, listingId int, toRemove int) (int, error) {
	stmt, err := db.Prepare(`
		SELECT seats FROM reservations
			WHERE passenger_id = ? AND listing_id = ?
	`)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()
	var seats int
	err = stmt.QueryRow(toRemove, listingId).Scan(&seats)
	if err != nil {
		return seats, err
	}
	return seats, nil
}

func CheckPost(db *sql.DB, userId int, r *http.Request, listingId int) error {
	if r.FormValue("a") != "" {
		passengerId, err := strconv.Atoi(r.FormValue("a"))
		if err != nil {
			return errors.New("Invalid")
		}
		pendingUser, err := getPendingUser(db, listingId, passengerId)
		if err != nil {
			return err
		}
		deleted, err := deleteFromQueue(db, userId, listingId, passengerId)
		if err != nil {
			return err
		}
		if deleted {
			err := addToReservation(db, userId, listingId, passengerId, pendingUser.Seats)
			if err != nil {
				return err
			}
			err = updateSeats(db, userId, listingId, (pendingUser.Seats * -1))
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
		deleted, err := deleteFromQueue(db, userId, listingId, remove)
		if err != nil {
			return err
		}
		if deleted == false {
			seats, err := findSeats(db, listingId, remove)
			if err != nil {
				return err
			}
			_, err = deleteFromReservations(db, userId, listingId, remove)
			if err != nil {
				return err
			}
			err = updateSeats(db, userId, listingId, seats)
			if err != nil {
				return err
			}
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

func SendMessage(db *sql.DB, sender int, receiver int, message string) error {
		stmt, err := db.Prepare(`
		INSERT INTO messages(sender, receiver, message)
			SELECT ? AS sender, ? AS receiver, ? AS message FROM dual
				WHERE EXISTS (
					SELECT listing_id
						FROM reservations
						WHERE (sender = ? AND receiver = ?)
					OR (receiver = ? AND sender = ?)
			) LIMIT 1;
		`)
	
	if err != nil {
		return err // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err := stmt.Exec(sender, receiver, message, seats, sender, receiver, receiver, sender)
	if err != nil {
		return err
	}
	return nil
}
