package gen

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/util"
)

type PasswordData struct {
	Hash string
	Session string
	Salt string
}

func UnusedUsername(db *sql.DB, username string) bool {
	results := make ([]listing, 0)

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
	rows, err := stmt.Query(o, d)
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

func hashPassword (password string) PasswordData{
	hashed, err := newFromPassword(password, 10)
	if err != nil{
		panic(err.Error())
	}
	d := PasswordData{hashed.hash, "123", hashed.salt}
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

	afterHash:= hashPassword(password)
	addUser.Exec(username, email, afterHash.Hash, afterHash.session, afterHash.salt)

}