package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"5sur/util"
	"errors"
)

func deleteUpdateSeats(db *sql.DB, user int) error {
	stmt, err := db.Prepare(`
		UPDATE listings AS l
		JOIN reservations AS r
			ON r.passenger_id = ?
		SET l.seats = l.seats + r.seats
		WHERE l.id = r.listing_id;
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}

func deleteAllAlerts(db *sql.DB, user int) error {
	stmt, err := db.Prepare(`
		DELETE FROM alerts
			WHERE user = ?
			OR ((category = "message" OR category = "rate") AND target_id = ?);
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user, user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}

func deleteEmailPrefs(db *sql.DB, user int) error {
	stmt, err := db.Prepare(`
		DELETE FROM email_pref
			WHERE user = ?;
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}

func deleteDeleteAlerts(db *sql.DB, user int) error {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT DISTINCT l.id, l.driver
			FROM listings AS l
			JOIN reservations AS r
				ON r.listing_id = l.id
			WHERE r.passenger_id = ?;
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var id, driver int
		err := rows.Scan(&id, &driver)
		if err != nil {
			return util.NewError(err, "Database error", 500)
		}
		err = CreateAlert(db, driver, "dropped", id)
		if err != nil { return err }
	}
	err = deleteAllAlerts(db, user)
	if err !=nil { return err }
	return nil
}

func DeleteAccount(db *sql.DB, user int) error {
	// We may have to update all this to transactions, but we'd need to upgrade to go1.4
	// Reason being, if the delete sql fails, then the seats were already updated
	err := deleteUpdateSeats(db, user)
	if err != nil { return err }

	err = deleteDeleteAlerts(db, user)
	if err != nil { return err }

	err = deleteEmailPrefs(db, user)
	if err != nil { return err }

	stmt, err := db.Prepare(`
		DELETE u, l, m, r, rq, rh
			FROM users AS u
			LEFT JOIN listings AS l ON l.driver = u.id
			LEFT JOIN messages AS m ON (m.sender = u.id OR m.receiver = u.id)
			LEFT JOIN reservations AS r ON (r.driver_id = u.id OR r.passenger_id = u.id)
			LEFT JOIN reservation_queue AS rq ON rq.passenger_id = u.id
			LEFT JOIN ride_history AS rh ON user_id = u.id
			WHERE u.id = ?;
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	res, err := stmt.Exec(user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	num, err := res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	if num == 0 {
		return util.NewError(errors.New("Delete function didn't delete anything"), "Database error", 500)
	}
	return nil
}