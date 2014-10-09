package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"code.google.com/p/go.crypto/bcrypt"
)

func UnusedUsername(db *sql.DB, username string) bool {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
	SELECT users.name
		FROM users
		WHERE users.name = ?;
		`)
	
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 19`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 27`)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return true
	}
	return false
}

func hashPassword (password string) []byte{
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil{
		panic(err.Error() + ` THE ERROR IS ON LINE 43`)
	}
	return hashed
}

func CreateUser(db *sql.DB, username string, password string, email string){
		// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		INSERT INTO users (name, email, password, session)
			VALUES (?, ?, ?, ?)
		`)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 56`)
	}
	_, err = stmt.Exec(username, email, hashPassword(password), "123")
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 60`)
	}
	/*
	rowCnt, err := res.RowsAffected()
	if err != nil {
		// Log the error
	}
	*/
}