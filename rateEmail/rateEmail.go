package main

import (
	"rateEmail/util"
	"database/sql"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
)

type rateEmail struct {
	id int
	email string
	user string
	target int
	emailPref bool
}

/** Human readable errors here do not need to be translated, or even explicilty stated **/

func checkRateAlerts(db *sql.DB) ([]rateEmail, error) {
	var results []rateEmail
	stmt, err := db.Prepare(`
		SELECT a.id, u2.email, u.name, a.target_id, e.rate
			FROM alerts AS a
				JOIN users AS u
				JOIN users AS u2
				JOIN email_pref AS e ON e.user = a.user
			WHERE a.category = "rate"
				AND u.id = a.target_id
				AND u2.id = a.user
				AND rate_email = 0;
	`)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	var result rateEmail

	for rows.Next() {
		err := rows.Scan(&result.id, &result.email, &result.user, &result.target, &result.emailPref)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}

		results = append(results, result)
	}
	return results, nil
}

func setRateEmail(db *sql.DB, id int) error {
	stmt, err := db.Prepare(`
		UPDATE alerts
			SET rate_email = 1
			WHERE id = ?;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func main() {
	db, err := util.OpenDb()
	if err != nil { 
		if myErr, ok := err.(util.MyError); ok {
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}
		}
		return
	}
	defer db.Close()

	rateEmails, err := checkRateAlerts(db)
	if err != nil { 
		if myErr, ok := err.(util.MyError); ok {
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}
		}
		return
	}

	subject := "5sur - Ahora puedes dar rating a un usuario"
	for _, rateEmail := range rateEmails {
		err = setRateEmail(db, rateEmail.id)
		if err != nil { 
			if myErr, ok := err.(util.MyError); ok {
				if myErr.LogError != nil {
					util.PrintLog(myErr)
				}
			}
		} else if rateEmail.emailPref {
			text := "Ahora puedes dar un rating a " + rateEmail.user + " por un viaje recién compartido"
			buttonText := "Dar puntaje"
			buttonLink := "https://5sur.com/rate?i=" + strconv.Itoa(rateEmail.target)
			body := util.EmailTemplate(text, buttonText, buttonLink)
			err = util.SendEmail(rateEmail.email, subject, body)
		}
	}
	if err != nil { 
		if myErr, ok := err.(util.MyError); ok {
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}
		}
		return
	}

}