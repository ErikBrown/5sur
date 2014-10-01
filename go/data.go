package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Listing struct {
	datePosting string
	dateLeaving string
	driver string
	picture string
	origin string
	destination string
	fee float32
}

func main() {
	// The db should be long lived. Do not recreate it unless accessing a different
	// database. Do not Open() and Close() from a short lived function, just pass in
	// the db object to the function
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// Defer Close() to be run at the end of main()
	defer db.Close()

	// sql.Open does not establish any connections to the database - To check if the
	// database is available and accessable, use sql.Ping()
	err = db.Ping()
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare("SELECT l.date_posting, l.date_leaving, u.name, u.picture, l.origin, l.destination, l.fee FROM listings AS l JOIN users AS u ON l.driver = u.id LIMIT ?")
	
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	results := make ([]Listing, 0) // Make struct to store everything

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(2)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		var temp Listing
		err := rows.Scan(&temp.datePosting, &temp.dateLeaving, &temp.driver, &temp.picture, &temp.origin, &temp.destination, &temp.fee)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		results = append(results, temp)
	}

	fmt.Printf("%v\n", results[1].driver)
}