package util

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/smtp"
	"net/http"
	"io/ioutil"
	"crypto/tls"
)

type EmailPrefs struct {
	User int
	Pending bool
	Dropped bool
	Message bool
	Accepted bool
	Removed bool
	Deleted bool
	Rate bool
}

func SendEmail(toAddress string, subject string, body string) error {
	from := "admin@5sur.com"
	to := toAddress

	// Setup message (need the carriage return \r before body)
	message := "From: " + from + "\r\n"
	message += "To: " + to + "\r\n"
	message += "Subject: " + subject + "\r\n"
	message += "Content-Type: text/html; charset=UTF-8" + "\r\n"
	message += "\r\n" + body

	// SMTP Server info
	user, err := ioutil.ReadFile("sesUser")
	if err != nil {
		return NewError(err, "Internal server error", 500)
	}
	password, err := ioutil.ReadFile("sesPassword")
	if err != nil {
		return NewError(err, "Internal server error", 500)
	}
	servername := "email-smtp.us-west-2.amazonaws.com:465"
	host := "email-smtp.us-west-2.amazonaws.com"
	auth := smtp.PlainAuth("", string(user[:]), string(password[:]), host)

	// TLS config
	tlsconfig := &tls.Config {
		ServerName: host,
	}

	conn, err := tls.Dial("tcp",servername,tlsconfig)
	if err != nil {
		return NewError(err, "Email authentication error", 500)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return NewError(err, "Email authentication error", 500)
	}
	defer client.Quit()

	// auth
	if err = client.Auth(auth); err != nil {
		return NewError(err, "Email authentication error", 500)
	}

	// to and from
	if err = client.Mail(from); err != nil {
		return NewError(err, "Email authentication error", 500)
	}

	// Can have multiple Rcpt calls
	if err = client.Rcpt(to); err != nil {
		return NewError(err, "Email authentication error", 500)
	}

	// Data
	dataWriter, err := client.Data()
	if err != nil {
		return NewError(err, "Email authentication error", 500)
	}
	defer dataWriter.Close()

	_, err = dataWriter.Write([]byte(message))
	if err != nil {
		return NewError(err, "Email authentication error", 500)
	}
	return nil
}

func SetEmailPref(db *sql.DB, r *http.Request, user int) error {
	prefs := EmailPrefs{User: user}
		if r.FormValue("pending") != "" {
			prefs.Pending = true
		}
		if r.FormValue("dropped") != "" {
			prefs.Dropped = true
		}
		if r.FormValue("message") != "" {
			prefs.Message = true
		}
		if r.FormValue("accepted") != "" {
			prefs.Accepted = true
		}
		if r.FormValue("removed") != "" {
			prefs.Removed = true
		}
		if r.FormValue("deleted") != "" {
			prefs.Deleted = true
		}
		if r.FormValue("rate") != "" {
			prefs.Rate = true
		}

	stmt, err := db.Prepare(`
		UPDATE email_pref
			SET pending = ?,
				dropped = ?,
				message = ?,
				accepted = ?,
				removed = ?,
				deleted = ?,
				rate = ?
			WHERE user = ?;
		`)
	if err != nil {
		return NewError(err, "Database error", 500)
	}

	defer stmt.Close()

	_, err = stmt.Exec(prefs.Pending, prefs.Dropped, prefs.Message, prefs.Accepted, prefs.Removed, prefs.Deleted, prefs.Rate, user)
	if err != nil {
		return NewError(err, "Database error", 500)
	}

	return nil
}

func ReturnEmailPref(db *sql.DB, user int) (EmailPrefs, error) {
	prefs := EmailPrefs{User: user}
	stmt, err := db.Prepare(`
		SELECT pending, dropped, message, accepted, removed, deleted, rate
			FROM email_pref
			WHERE user = ?;
	`)

	if err != nil {
		return prefs, NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	err = stmt.QueryRow(user).Scan(&prefs.Pending, &prefs.Dropped, &prefs.Message, &prefs.Accepted, &prefs.Removed, &prefs.Deleted, &prefs.Rate)
	if err != nil {
		return prefs, nil
	}

	return prefs, nil
}

func CreateEmailPrefs(db *sql.DB, user int64) error {
	stmt, err := db.Prepare(`
		INSERT INTO email_pref (user, pending, dropped, message, accepted, removed, deleted, rate)
			VALUES(?, true, true, true, true, true, true, true);
		`)
	if err != nil {
		return NewError(err, "Database error", 500)
	}

	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return NewError(err, "Database error", 500)
	}

	return nil
}