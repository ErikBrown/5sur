package util

import (
	"io/ioutil"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func OpenDb() (*sql.DB, error) {
	user, err := ioutil.ReadFile("dbUser")
	if err != nil {
		return &sql.DB{}, NewError(err, "Error de servidor", 500)
	}

	password, err := ioutil.ReadFile("dbPassword")
	if err != nil {
		return &sql.DB{}, NewError(err, "Error de servidor", 500)
	}

	db, err := sql.Open("mysql", string(user[:]) + ":" + string(password[:]) + "@/rideshare")
	if err != nil {
		return db, NewError(err, "Conexión fallada", 500)
	}
	err = db.Ping()
	if err != nil {
		return db, NewError(err, "Conexión fallada", 500)
	}
	return db, nil
}