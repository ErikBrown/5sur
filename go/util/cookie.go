package util

import (
	"time"
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func CreateCookie(u string, db *sql.DB) (http.Cookie, error) {
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

	err := updateSession(randValue, u, db)
	if err != nil {
		return http.Cookie{}, err
	}
	return authCookie, nil
}

func DeleteCookie() http.Cookie {
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
	return expiredCookie
}

func updateSession(v string, u string, db *sql.DB) error {
	// To save CPU cycles we'll use sha256; Bcrypt is an intentionally slow hash.
	// We don't even need that secure of a hash function since our session ID is 
	// sufficiently random and long.
	hashed := sha256.New()
	hashed.Write([]byte(v))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmt, err := db.Prepare(`UPDATE users SET session = ? WHERE name = ?`)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	res, err := stmt.Exec(hashedStr, u)
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return NewError(err, "Database error", 500)
	}
	return nil
}

func CheckCookie(r *http.Request, db *sql.DB) (string, int, error) {
	sessionID, err := r.Cookie("RideChile")
	if err != nil {
		return "", 0, nil // No cookie
	}
	n, i, err := checkSession(sessionID.Value, db)
	if err != nil {return "", 0, err}

	return n, i, nil
}

func checkSession(s string, db *sql.DB) (string, int, error){
	hashed := sha256.New()
	hashed.Write([]byte(s))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmt, err := db.Prepare(`
	SELECT users.name, users.id
		FROM users
		WHERE users.session = ?
		`)
	
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