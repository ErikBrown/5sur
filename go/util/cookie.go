package util

import (
	"time"
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func CreateCookie(u string, db *sql.DB, app bool) (http.Cookie, error) {
	alphaNum := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv")
	randValue := ""
	for i := 0; i < 32; i++ {
		num, err := RandKey(58)
		if err != nil {return http.Cookie{}, err}
		randValue = randValue + string(alphaNum[num])
	}

	authCookie := http.Cookie{
		Name: "RideChile",
		Value: randValue,
		Path: "/",
		Domain: "5sur.com", // Add domain name in the future
		Expires: time.Now().AddDate(0,1,0), // One month from now
		Secure: true, // SSL only
		HttpOnly: true, // HTTP(S) only
	}

	err := updateSession(randValue, u, db, app)
	if err != nil {
		return http.Cookie{}, err
	}
	return authCookie, nil
}

func DeleteCookie(db *sql.DB, userId int, app bool) (error, http.Cookie) {
	expiredCookie := http.Cookie{
		Name: "RideChile",
		Value: "",
		Path: "/",
		Domain: "5sur.com", // Add domain name in the future
		Expires: time.Now().Add(-1000), // Expire cookie
		MaxAge: -1,
		Secure: true, // SSL only
		HttpOnly: true, // HTTP(S) only
	}
	err := deleteAuthToken(db, userId, app)
	if err != nil {
		return err, expiredCookie
	}
	return nil, expiredCookie
}

func deleteAuthToken(db *sql.DB, userId int, app bool) error {
	stmtText := ""
	if app {
		stmtText = `UPDATE users SET ios_session = "" WHERE id = ?`
	} else {
		stmtText = `UPDATE users SET session = "" WHERE id = ?`
	}
	stmt, err := db.Prepare(stmtText)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userId)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	return nil
}

func updateSession(v string, u string, db *sql.DB, app bool) error {
	// To save CPU cycles we'll use sha256; Bcrypt is an intentionally slow hash.
	// We don't even need that secure of a hash function since our session ID is 
	// sufficiently random and long.
	hashed := sha256.New()
	hashed.Write([]byte(v))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmtText := ""
	if app {
		stmtText = `UPDATE users SET ios_session = ? WHERE name = ?`
	} else {
		stmtText = `UPDATE users SET session = ? WHERE name = ?`
	}
	stmt, err := db.Prepare(stmtText)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(hashedStr, u)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	return nil
}

func CheckAppCookie(r *http.Request, db *sql.DB) (string, int, error) {
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		return "", 0, nil // No cookie
	}
	n, i, err := checkSession(sessionID.Value, true, db)
	if err != nil {return "", 0, err}

	return n, i, nil
}

func CheckCookie(r *http.Request, db *sql.DB) (string, int, error) {
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		return "", 0, nil // No cookie
	}
	n, i, err := checkSession(sessionID.Value, false, db)
	if err != nil {return "", 0, err}

	return n, i, nil
}

func checkSession(s string, app bool, db *sql.DB) (string, int, error){
	hashed := sha256.New()
	hashed.Write([]byte(s))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmtText := ""
	if app {
		stmtText = `
	SELECT users.name, users.id
		FROM users
		WHERE users.ios_session = ?
		AND users.ios_session != "";
		`
	} else {
		stmtText = `
	SELECT users.name, users.id
		FROM users
		WHERE users.session = ?
		AND users.session != "";
		`
	}
	stmt, err := db.Prepare(stmtText)
	
	if err != nil {
		return "", 0, NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(hashedStr)
	if err != nil {
		return "", 0, NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()

	name := ""
	var id int

	for rows.Next() {
		err := rows.Scan(&name, &id)
		if err != nil {
			return "", 0, NewError(err, "Database error", 500)
		}
		
	}
	return name, id, nil
}