package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"code.google.com/p/go.crypto/bcrypt"
	"bytes"
)

func UnusedUsername(db *sql.DB, username string) bool {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT users.name
			FROM users
			WHERE l.origin = ?
		`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return false
	}
	return true
}

func hashPassword (password string) string{
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil{
		panic(err.Error())
	}
	n := bytes.Index(hashed, []byte{0})
	return string(hashed[:n])
}

func CreateUser(db *sql.DB, username string, password string, email string){
		// Always prepare queries to be used multiple times. The parameter placehold is ?
	addUser, err := db.Prepare(`
		INSERT INTO users (name, email, password, session, salt)
			VALUES (?, ?, ?, ?, ?)
		`)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer addUser.Close()
	addUser.Exec(username, email, hashPassword(password), "123")

}