package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"5sur/util"
	"unicode/utf8"
)

func CreateReservation(db *sql.DB, userId int, listingId int, seats int, message string) error {
	ride, err := ReturnIndividualListing(db, listingId)
	if err != nil {
		return err
	}

	if userId == ride.Driver {
		return util.NewError(nil, "No puedes registrarte por tu propio viaje", 400)
	}

	if seats > ride.Seats {
		return util.NewError(nil, "No hay cupos disponibles", 400)
	}

	if seats <= 0 {
		return util.NewError(nil, "Tienes que registrarte por la menos un asiento", 400)
	}

	if utf8.RuneCountInString(message) > 200 {
		return util.NewError(nil, "Mensaje demasiado largo. Max caracteres 200", 400)
	}
	
	err = validReservation(db, userId, listingId, ride.Timestamp)
	if err != nil {
		return err
	}

	err = makeReservation(db, listingId, seats, userId, message)
	if err != nil {
		return err
	}
	err = CreateAlert(db, ride.Driver, "pending", listingId)
	if err != nil { return err }
	return nil
}

func validReservation(db *sql.DB, userId int, listingId int, date string) error {
	stmt, err := db.Prepare(`
		SELECT r.id
			FROM reservation_queue as r
			WHERE r.listing_id = ? AND r.passenger_id = ?
	`)
	
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	t := ""
	err = stmt.QueryRow(listingId, userId).Scan(&t)
	if err == nil {
		return util.NewError(nil, "Ya estás en la lista de reservaciones ", 400)
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
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt2.Close()

	rows, err := stmt2.Query(userId, date, date)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer rows.Close()

	results := make ([]int, 0)
	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp int
		err := rows.Scan(&temp)
		if err != nil {
			return util.NewError(err, "Error de la base de datos", 500)
		}
		results = append(results, temp)
	}
	for _, v := range results {
		if v == listingId {
			return util.NewError(nil, "Ya estas registrado por este viaje", 400)
		}
	}

	if len(results) != 0 {
		return util.NewError(nil, "Ya estas registrado por un viaje en esta fecha", 400)
	}

	return nil
}

func CheckReservationQueue(db *sql.DB, listingId int) (bool, error) {
	stmt, err := db.Prepare(`
		SELECT * FROM reservation_queue 
			WHERE listing_id = ?
		`)
	if err != nil {
		return false, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()
	rows, err := stmt.Query(listingId)
	if err != nil {
		return false, util.NewError(err, "Error de la base de datos", 500)
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
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(listingId, seats, userId, message)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}
	return nil
}