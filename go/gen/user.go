package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func ReturnUserInfo(db *sql.DB, u string) User {
	var results User

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
	SELECT u.name, u.picture, u.created, u.positive_ratings, u.negative_ratings
		FROM users as u
		WHERE u.name = ?
	`)
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(u)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&results.Name, &results.Picture, &results.Created, &results.RatingPositive, &results.RatingNegative)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
	}
	
	return results
}