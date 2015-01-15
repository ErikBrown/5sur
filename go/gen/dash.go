package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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

type DashMessages struct {
	Id int
	Name string
	Picture string
	Count int
}

type DashReservation struct {
	ListingId int
	Day string
	Month string
	Time string
	Origin string
	Destination string
	Seats int
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
	Day string
	Month string
	Time string
	Origin string
	Destination string
	Seats int
	Fee string
	ListingId int
	DriverId int
	DriverName string
	DriverPicture string
}

type MessageThread struct {
	UserId int
	Name string
	Picture string
	Messages []SpecificMessage
}

type SpecificMessage struct {
	Id int
	Sent bool
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
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}

		// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		date := ""
		var temp DashListing
		err := rows.Scan(&date, &temp.Origin, &temp.Destination, &temp.ListingId, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		convertedDate, err := util.PrettyDate(date, false)
		if err != nil { return results, err }
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
		SELECT l.id, l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN cities as c ON l.origin = c.id
			LEFT JOIN cities as c2 ON l.destination = c2.id
			JOIN reservations as r ON l.id = r.listing_id
			WHERE r.passenger_id = ?
			ORDER BY l.date_leaving;
		`)

	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}

		// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		temp := DashReservation{}
		date := ""
		err := rows.Scan(&temp.ListingId, &date, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		convertedDate, err := util.PrettyDate(date, false)
		if err != nil { return results, err }
		temp.Day = convertedDate.Day
		temp.Month = convertedDate.Month
		temp.Time = convertedDate.Time
		results = append(results, temp)
	}

	return results, nil
}

func GetDashMessages(db *sql.DB, userId int) ([]DashMessages, error) {
	var results []DashMessages

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT m.sender, u.name, u.picture, SUM(IF(m.opened = 0, 1, 0))
			FROM messages as m 
			JOIN users AS u 
				ON u.id = m.sender 
			WHERE m.receiver = ?
			GROUP BY m.sender;
		`)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}

	for rows.Next() {
		var temp DashMessages
		err := rows.Scan(&temp.Id, &temp.Name, &temp.Picture, &temp.Count)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		results = append(results, temp)
	}

	return results, nil
}

// The naming is poor, this actually returns all messages sent between you and whoever else
func SpecificDashMessage(db *sql.DB, messages []DashMessages, recipient int, userId int) (MessageThread, error) {
	found := false
	message := MessageThread{}
	for i := range messages{
		if messages[i].Id == recipient {
			message.UserId = messages[i].Id
			message.Name = messages[i].Name
			message.Picture = messages[i].Picture
			found = true
			break
		}
	}
	if !found {
		return MessageThread{}, util.NewError(nil, "Could not find specific message", 400)
	}
	var err error
	message.Messages, err = getMessages(db, recipient, userId)
	if err != nil {
		return MessageThread{}, err
	}
	return message, nil
}

func getMessages(db *sql.DB, recipient int, userId int) ([]SpecificMessage, error) {
	results := make ([]SpecificMessage, 0)
	stmt, err := db.Prepare(`
		SELECT m.id, m.sender, m.date, m.message 
			FROM messages AS m 
			WHERE (receiver = ? AND sender = ?) 
				OR (receiver = ? AND sender = ?);
		`)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(recipient, userId, userId, recipient)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}

	for rows.Next() {
		temp := SpecificMessage{}
		sender := 0
		var s sql.NullString
		err := rows.Scan(&temp.Id, &sender, &temp.Date, &s)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		// Check for null value
		if s.Valid {
			temp.Message = s.String
		}

		if sender == userId {
			temp.Sent = true
		} else {
			temp.Sent = false
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
		return SpecificListing{}, util.NewError(nil, "Could not find specific listing", 400)
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
			results.Day = reservations[i].Day
			results.Month = reservations[i].Month
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
		return Reservation{}, util.NewError(nil, "Could not find specific reservation", 400)
	}

	stmt, err := db.Prepare(`
		SELECT u.id, u.name, u.picture
			FROM users as u
			LEFT JOIN reservations as r ON u.id = r.driver_id
			WHERE r.listing_id = ?
		`)
	if err != nil {
		return Reservation{}, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		return Reservation{}, util.NewError(err, "Database error", 500)
	}

	for rows.Next() {
		err := rows.Scan(&results.DriverId, &results.DriverName, &results.DriverPicture)
		if err != nil {
		return Reservation{}, util.NewError(err, "Database error", 500)
		}
	}

	return results, nil
}

func getPendingUsers(db *sql.DB, listingId int) ([]PendingUser, error) {
	var results []PendingUser
	stmt, err := db.Prepare(`
		SELECT r.message, u.id, u.name, u.picture, r.seats
			FROM reservation_queue as r
			JOIN users AS u ON r.passenger_id = u.id
			WHERE r.listing_id = ?;
	`)
	
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var temp PendingUser
		err := rows.Scan(&temp.Message, &temp.Id, &temp.Name, &temp.Picture, &temp.Seats)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
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
		return PendingUser{}, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	pendingUser := PendingUser{}
	err = stmt.QueryRow(listingId, pendingUserId).Scan(&pendingUser.Message, &pendingUser.Id, &pendingUser.Name, &pendingUser.Picture, &pendingUser.Seats)
	if err != nil {
		return pendingUser, util.NewError(err, "User does not exist", 400)
	}
	return pendingUser, nil
}

func getRegisteredUsers(db *sql.DB, listingId int) ([]RegisteredUser, error) {
	var results []RegisteredUser
	stmt, err := db.Prepare(`
		SELECT u.id, u.name, u.picture, r.seats
			FROM reservations as r
			JOIN users AS u ON r.passenger_id = u.id
			WHERE r.listing_id = ?;
	`)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(listingId)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var temp RegisteredUser
		err := rows.Scan( &temp.Id, &temp.Name, &temp.Picture, &temp.Seats)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
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
		return false, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	affected, err := stmt.Exec(passenger_id, userId, listingId, listingId)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	rowsDeleted, err := affected.RowsAffected()
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	if rowsDeleted == 0{
		return false, nil
	}
	return true, nil
}

func deleteFromReservations(db *sql.DB, driverId int, listingId int, passengerId int) (bool, error) {
	stmt, err := db.Prepare(`
		DELETE FROM reservations
			WHERE driver_id = ?
				AND listing_id = ?
				AND passenger_id = ?
		`)
	
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	affected, err := stmt.Exec(driverId, listingId, passengerId)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	rowsDeleted, err := affected.RowsAffected()
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
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
				WHERE NOT EXISTS (
					SELECT listing_id
						FROM reservations
						WHERE listing_id = ?
						AND driver_id = ?
						AND passenger_id = ?
				) LIMIT 1;
		`)
	
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(listingId, userId, passengerId, seats, listingId, userId, passengerId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
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
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()
	_, err = stmt.Exec(seats, listingId, userId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func findSeats(db *sql.DB, listingId int, toRemove int) (int, error) {
	stmt, err := db.Prepare(`
		SELECT seats FROM reservations
			WHERE passenger_id = ? AND listing_id = ?
	`)
	if err != nil {
		return 0, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()
	var seats int
	err = stmt.QueryRow(toRemove, listingId).Scan(&seats)
	if err != nil {
		return seats, util.NewError(err, "Database error", 500)
	}
	return seats, nil
}

func DeleteListing(db *sql.DB, userId int, listingId int) ([]RegisteredUser, error) {
	registeredUsers, err := getRegisteredUsers(db, listingId)
	if err != nil {return registeredUsers, err}
	stmt, err := db.Prepare(`
		DELETE l, r, rq
			FROM listings AS l
			LEFT JOIN reservations AS r ON r.listing_id = l.id
			LEFT JOIN reservation_queue AS rq ON rq.listing_id = l.id
			WHERE l.driver = ?
				AND l.id = ?
	`)
	if err != nil {
		return registeredUsers, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(userId, listingId)
	if err != nil {
		return registeredUsers, util.NewError(err, "Database error", 500)
	}
	return registeredUsers, nil
}

// userId in this case is passengerId because this is for a user removing themself from the reservation list
// or messenging the driver of the ride.
func CheckReservePost(db *sql.DB, userId int, r *http.Request, listingId int) (string, error) {
	if r.FormValue("r") != "" {
		// Handle deleting this reservation
		driverId, err := strconv.Atoi(r.FormValue("r"))
		if err != nil {
			return "", util.NewError(nil, "Invalid user", 400)
		}
		seats, err := findSeats(db, listingId, userId)
		if err != nil {
			return "", err
		}
		deleted, err := deleteFromReservations(db, driverId, listingId, userId)
		if err != nil {
			return "", err
		}
		if deleted {
			err = updateSeats(db, driverId, listingId, seats)
			if err != nil {
				return "", err
			}
			err = CreateAlert(db, driverId, "dropped", listingId)
			if err != nil { return "", err }
		}
		return "https://5sur.com/dashboard/reservations", nil
	}
	if r.FormValue("m") != "" {
		// We are messenging the user with id equal to the post request data.
		return "https://5sur.com/dashboard/messages", nil
	}
	return "", nil
}

func CheckPost(db *sql.DB, userId int, r *http.Request, listingId int) error {
	if r.FormValue("a") != "" {
		passengerId, err := strconv.Atoi(r.FormValue("a"))
		if err != nil {
			return util.NewError(nil, "Invalid passenger", 400)
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
			pendingUsers, err := getPendingUsers(db, listingId)
			if err != nil { return err }
			if len(pendingUsers) == 0 {
				err = DeleteAlert(db, userId, "pending", listingId)
				if err != nil { return err }
			}
			err = CreateAlert(db, passengerId, "accepted", listingId)
			if err != nil { return err }
		}
		return nil
	}
	if r.FormValue("r") != "" {
		passengerId, err := strconv.Atoi(r.FormValue("r"))
		if err != nil {
			return util.NewError(nil, "Invalid passenger", 400)
		}
		deleted, err := deleteFromQueue(db, userId, listingId, passengerId)
		if err != nil {
			return err
		}
		if deleted == false {
			seats, err := findSeats(db, listingId, passengerId)
			if err != nil {
				return err
			}
			err = DeleteAlert(db, passengerId, "accepted", listingId)
			if err != nil { return err }
			_, err = deleteFromReservations(db, userId, listingId, passengerId)
			if err != nil {
				return err
			}
			err = updateSeats(db, userId, listingId, seats)
			if err != nil {
				return err
			}
			err = CreateAlert(db, passengerId, "removed", listingId)
			if err != nil { return err }
		} else {
			pendingUsers, err := getPendingUsers(db, listingId)
			if err != nil { return err }
			if len(pendingUsers) == 0 {
				err = DeleteAlert(db, userId, "pending", listingId)
				if err != nil { return err }
			}
		}
		return nil
	}
	if r.FormValue("m") != "" {
		_, err := strconv.Atoi(r.FormValue("m"))
		if err != nil {
			return util.NewError(nil, "Invalid passenger", 400)
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
						FROM reservations as r
						WHERE (r.driver_id = ? AND r.passenger_id = ?)
						OR (r.passenger_id = ? AND r.driver_id = ?)
				) LIMIT 1;
		`)
	
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	res, err := stmt.Exec(sender, receiver, message, sender, receiver, sender, receiver)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	if rowCnt == 0 {
		return util.NewError(err, "You do not have permissions to message this person", 400)
	}

	err = CreateAlert(db, receiver, "message", sender)
	if err != nil { return err }

	return nil
}

func SetMessagesClosed(db *sql.DB, sender int, receiver int) error {
	stmt, err := db.Prepare(`
		UPDATE messages SET opened = 1
			WHERE sender = ?
			AND receiver = ?
		`)
	
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	res, err := stmt.Exec(sender, receiver)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}