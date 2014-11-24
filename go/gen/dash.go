package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DashListing struct {
	Day string
	Month string
	Origin int
	Destination int
	Alert bool
	ListingId int
}

func GetDashListings(db *sql.DB, userId int) ([]DashListing, error) {
	results := make ([]DashListing, 0)

		// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		SELECT l.date_leaving, l.origin, l.destination, l.id
			FROM listings AS l
			WHERE l.date_leaving >= NOW() AND l.driver = ?
			ORDER BY l.date_leaving
			LIMIT 25
		`)

	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		panic(err.Error()) // Have a proper error in production
	}

		// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		date := ""
		var temp DashListing
		err := rows.Scan(&date, &temp.Origin, &temp.Destination, &temp.ListingId)
		if err != nil {
			panic(err.Error()) // Have a proper error in production
		}
		convertedDate := util.PrettyDate(date, false)
		temp.Day = convertedDate.Day
		temp.Month = convertedDate.Month
		temp.Alert, err = CheckReservationQueue(db, temp.ListingId)
		if err != nil {
			return results, err
		}
		// Also find if there are any new messages.
		results = append(results, temp)
	}

	return results, nil

}