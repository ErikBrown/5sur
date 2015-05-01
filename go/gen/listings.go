package gen

import (
	"database/sql"
	"5sur/util"
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
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query()
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp City
		err := rows.Scan(&temp.Id, &temp.Name)
		if err != nil {
			return results, util.NewError(err, "Error de la base de datos", 500)
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
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(date_leaving, date_leaving, id)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer rows.Close()

	for rows.Next() {
		return util.NewError(nil, "Ya tienes un viaje en esta fecha", 400)
	}
	return nil
}

func ReturnAllListings(db *sql.DB) ([]Listing, error) {
	results := make ([]Listing, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.name, u.custom_picture, (u.positive_ratings - u.negative_ratings), l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.seats > 0
				AND l.date_leaving > NOW()
			ORDER BY l.date_leaving
			LIMIT 50
		`)
	
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var temp Listing
		customPicture := false
		name := ""
		err := rows.Scan(&temp.Id, &temp.Driver, &name, &customPicture, &temp.Rating, &temp.Timestamp, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, util.NewError(err, "Error de la base de datos", 500)
		}
		prettyTime, err := util.PrettyDate(temp.Timestamp, false)
		if err != nil { return results, err }
		temp.Date = prettyTime.Month + " " + prettyTime.Day
		temp.Time = prettyTime.Time

		if customPicture {
			temp.Picture = "https://5sur.com/images/" + name + "_50.png"
		} else {
			temp.Picture = "https://5sur.com/default_50.png"
		}
		results = append(results, temp)
	}
	return results, nil
}

func ReturnListings(db *sql.DB, o int, d int, t string) ([]Listing, error) {
	results := make ([]Listing, 0)

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.name, u.custom_picture, (u.positive_ratings - u.negative_ratings), l.date_leaving, c.name, c2.name, l.seats, l.fee
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
			LIMIT 50
		`)
	
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(o, d, t)
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp Listing
		customPicture := false
		name := ""
		err := rows.Scan(&temp.Id, &temp.Driver, &name, &customPicture, &temp.Rating, &temp.Timestamp, &temp.Origin, &temp.Destination, &temp.Seats, &temp.Fee)
		if err != nil {
			return results, util.NewError(err, "Error de la base de datos", 500)
		}
		prettyTime, err := util.PrettyDate(temp.Timestamp, false)
		if err != nil { return results, err }
		temp.Date = prettyTime.Month + " " + prettyTime.Day
		temp.Time = prettyTime.Time

		if customPicture {
			temp.Picture = "https://5sur.com/images/" + name + "_50.png"
		} else {
			temp.Picture = "https://5sur.com/default_50.png"
		}
		results = append(results, temp)
	}
	return results, nil
}

func CreateListing(db *sql.DB, date_leaving string, driver int, origin int, destination int, seats int, fee float64) (int64, error) {
	// This needs to take in account the hour/minute!!! Concatinate the form values! CHANGE THIS
	listingTotal, err := checkListingTotal(db, driver)
	if err != nil { return 0, err }

	if listingTotal > 20 {
		return 0, util.NewError(err, "Tienes demasiados viajes existentes (max 20)", 400)
	}

	err = checkNearbyListings(db, date_leaving, driver)
	if err !=nil {
		return 0, err
	}

	stmt, err := db.Prepare(`
		INSERT INTO listings (date_leaving,driver,origin,destination,seats,fee)
			SELECT ? AS date_leaving, ? AS driver, ? AS origin, ? AS destination, ? AS seats, ? AS fee FROM dual
			WHERE 2 = (
				SELECT COUNT(*)
					FROM cities
					WHERE id = ?
					OR id = ?
				) LIMIT 1;
		`)
	if err != nil {
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	res, err := stmt.Exec(date_leaving, driver, origin, destination, seats, fee, origin, destination)
	if err != nil {
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}

	if rowCnt != 1 { // Invalid city id
		return 0, util.NewError(nil, "Parámetros inválidos", 400)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}
	
	return lastId, nil
}

func checkListingTotal(db *sql.DB, driver int) (int, error) {
	stmt, err := db.Prepare(`
		SELECT COUNT(*)
			FROM listings
			WHERE driver = ?
			AND date_leaving < NOW()
		`)
	
	if err != nil {
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	listingTotal := 0
	err = stmt.QueryRow(driver).Scan(&listingTotal)
	if err != nil {
		// This should never happen
		return 0, util.NewError(err, "Error de la base de datos", 500)
	}

	return listingTotal, nil

}

func ReturnIndividualListing(db *sql.DB, id int) (Listing, error) {
	result := Listing{}
	stmt, err := db.Prepare(`
		SELECT l.id, u.id, u.name, u.custom_picture, (u.positive_ratings - u.negative_ratings), l.date_leaving, c.name, c2.name, l.seats, l.fee
			FROM listings AS l
			JOIN users AS u ON l.driver = u.id
			JOIN cities AS c ON l.origin = c.id
			LEFT JOIN cities AS c2 ON l.destination = c2.id
			WHERE l.id = ?
		`)
	
	if err != nil {
		return Listing{}, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	customPicture := false
	name := ""
	err = stmt.QueryRow(id).Scan(&result.Id, &result.Driver, &name, &customPicture, &result.Rating, &result.Timestamp, &result.Origin, &result.Destination, &result.Seats, &result.Fee)
	if err != nil {
		return Listing{}, util.NewError(nil, "Viaje no existe", 400)
	}
	prettyTime, err := util.PrettyDate(result.Timestamp, false)
	if err != nil { return result, err }
	result.Date = prettyTime.Month + " " + prettyTime.Day
	result.Time = prettyTime.Time


	if customPicture {
		result.Picture = "https://5sur.com/images/" + name + "_50.png"
	} else {
		result.Picture = "https://5sur.com/default_50.png"
	}

	return result, nil
}