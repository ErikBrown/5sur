package gen

import (
	"5sur/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id int
	Name string
	Picture string
	Rating int
	RidesGivenTotal int
	RidesTakenTotal int
	RidesGiven map[string]int
	RidesTaken map[string]int
	Comments []Comment
}

type Comment struct {
	Positive bool
	Date string
	Text string
}

type RateParams struct {
	User int
	Positive bool
	Comment string
	Public bool
}

func ReturnUserInfo(db *sql.DB, u interface{}) (User, error) {
	var result User

	stmtText := ""

	switch u.(type) {
		case string:
			stmtText = `
				SELECT * FROM 
					(SELECT u.id, u.name, u.custom_picture, (positive_ratings - negative_ratings), sum(r.rides_given), sum(r.rides_taken) 
						FROM users AS u 
						LEFT JOIN ride_history AS r 
							ON r.user_id = u.id 
							WHERE u.name = ?)
					AS s 
					WHERE s.id IS NOT NULL;
				`
		case int:
			stmtText = `
				SELECT * FROM 
					(SELECT u.id, u.name, u.custom_picture, (positive_ratings - negative_ratings), sum(r.rides_given), sum(r.rides_taken)
						FROM users AS u 
						LEFT JOIN ride_history AS r 
							ON r.user_id = u.id 
							WHERE u.id = ?)
					AS s 
					WHERE s.id IS NOT NULL;
				`
	}
	stmt, err := db.Prepare(stmtText)
	
	if err != nil {
		return result, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	var given, taken sql.NullInt64

	customPicture := false
	err = stmt.QueryRow(u).Scan(&result.Id, &result.Name, &customPicture, &result.Rating, &given, &taken)
	if err != nil {
		return result, util.NewError(nil, "User does not exist", 404)
	}

	if customPicture {
		result.Picture = "https://5sur.com/images/" + result.Name + ".png"
	} else {
		result.Picture = "https://5sur.com/default.png"
	}

	if given.Valid {
		result.RidesGivenTotal = int(given.Int64)
	}

	if taken.Valid {
		result.RidesTakenTotal = int(taken.Int64)
	}

	// Ride history
	result.RidesGiven, result.RidesTaken, err = returnRideHistory(db, result.Id)
	if err != nil { return result, err }

	result.Comments, err = returnUserComments(db, result.Id)
	if err != nil { return result, err }

	return result, nil
}

func returnRideHistory(db *sql.DB, u int) (map[string]int, map[string]int, error) {
	result := make(map[string]int)
	result2 := make(map[string]int)


	stmt, err := db.Prepare(`
		SELECT year, sum(rides_given), sum(rides_taken)
			FROM ride_history 
			WHERE user_id = ? 
			GROUP BY year;
	`)

	if err != nil {
		return result, result2, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(u)
	if err != nil {
		return result, result2, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		year := ""
		var given, taken int
		err := rows.Scan(&year, &given, &taken)
		if err != nil {
			return result, result2, util.NewError(err, "Database error", 500)
		}
		result[year] = given
		result2[year] = taken
	}
	
	return result, result2, nil
}

func returnUserComments(db *sql.DB, u int) ([]Comment, error) {
	results := make ([]Comment, 0)

	stmt, err := db.Prepare(`
		SELECT positive, DATE_FORMAT(date,'%d/%m/%Y'), comment
			FROM comments 
			WHERE user = ? 
			AND public = true;
	`)

	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(u)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(&comment.Positive, &comment.Date, &comment.Text)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		results = append(results, comment)
	}
	
	return results, nil
}

func SubmitRating(db *sql.DB, commenter int, user int, positive bool, comment string, public bool) error {
	err := duplicateRating(db, user, commenter)
	if err != nil { return err }

	stmt, err := db.Prepare(`
		INSERT INTO comments (user, commenter, comment, positive, public)
			SELECT ? AS user, ? AS commenter, ? AS comment, ? AS positive, ? AS public
				FROM dual
				WHERE EXISTS (
					SELECT r.listing_id
						FROM reservations AS r
						JOIN listings AS l on r.listing_id = l.id
							WHERE ((r.driver_id = ? AND r.passenger_id = ?)
							OR (r.passenger_id = ? AND r.driver_id= ?))
							AND l.date_leaving < DATE_SUB(NOW(), INTERVAL 23 HOUR)
				) LIMIT 1;
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	res, err := stmt.Exec(user, commenter, comment, positive, public, commenter, user, commenter, user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	if rowCnt != 1 {
		return util.NewError(nil, "You cannot rate this user. You either have no past transactions with this user or one week has passed since last transaction", 400)
	}

	err = updateRatingScore(db, user, positive)
	if err != nil { return err }

	return nil
}

func updateRatingScore(db *sql.DB, user int, positive bool) error {
	stmtText := ""
	if positive {
		stmtText = "UPDATE users SET positive_ratings = positive_ratings + 1 WHERE id = ?;"
	} else {
		stmtText = "UPDATE users SET negative_ratings = negative_ratings + 1 WHERE id = ?;"
	}
	stmt, err := db.Prepare(stmtText)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}

func duplicateRating(db *sql.DB, user int, commenter int) error {
	stmt, err := db.Prepare(`
		SELECT id FROM comments WHERE user = ? AND commenter = ?;
		`)
	
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	id := 0
	err = stmt.QueryRow(user, commenter).Scan(&id)
	if err != nil {
		return nil
	}

	return util.NewError(nil, "You have already left a rating for this user", 400)
}

func ReturnUserPicture(db *sql.DB, user int, size string) (string, error) {
	picture := ""
	stmt, err := db.Prepare(`
		SELECT custom_picture, name FROM users WHERE id = ?;
		`)
	
	if err != nil {
		return picture, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	customPicture := false
	name := ""
	err = stmt.QueryRow(user).Scan(&customPicture, &name)
	if err != nil {
		return picture, util.NewError(err, "User not found", 500)
	}

	sizeSuffix := ""
	if size != "100" {
		sizeSuffix = "_" + size
	}

	if customPicture {
		picture = "https://5sur.com/images/" + name + sizeSuffix + ".png"
	} else {
		picture = "https://5sur.com/default" + sizeSuffix + ".png"
	}

	return picture, nil
}