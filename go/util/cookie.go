package util

import (
	"time"
	"net/http"
	"encoding/hex"
	"crypto/sha256"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func CreateCookie(u string, db *sql.DB) http.Cookie {
	alphaNum := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv")
	randValue := ""
	for i := 0; i < 32; i++ {
		randValue = randValue + string(alphaNum[RandKey(58)])
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

	updateSession(randValue, u, db)

	return authCookie
}

func updateSession(v string, u string, db *sql.DB) {
	// To save CPU cycles we'll use sha256; Bcrypt is an intentionally slow hash.
	// We don't even need that secure of a hash function since our session ID is 
	// sufficiently random and long.
	hashed := sha256.New()
	hashed.Write([]byte(v))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmt, err := db.Prepare(`UPDATE users SET session = ? WHERE name = ?`)
	defer stmt.Close()

	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 46`)
	}
	res, err := stmt.Exec(hashedStr, u)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 50`)
	}
	_, err = res.RowsAffected()
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 54`)
	}
}

func CheckCookie(session string, db *sql.DB) (string, int) {
	if session == "" {
		return "", 0
	}
	n, i := checkSession(session, db)

	if n != "" { // Super super temporary
		// You're logged in!
		return n, i
	} else {
		// Session ID is not valid
	}
	return n, i
}

func checkSession(s string, db *sql.DB) (string, int){
	hashed := sha256.New()
	hashed.Write([]byte(s))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmt, err := db.Prepare(`
	SELECT users.name, users.id
		FROM users
		WHERE users.session = ?
		`)
	
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 78`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(hashedStr)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 86`)
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
			panic(err.Error() + ` THE ERROR IS ON LINE 99`)
		}
		
	}
	return name, id
}