package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Alerts struct {
	Category string
	TargetId int
	Content string
}

func GetAlerts(db *sql.DB, user int) ([]Alerts, error) {
	var results []Alerts
	stmt, err := db.Prepare(`
		SELECT category, target_id, content
			FROM alerts
			WHERE user = ?;
	`)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user)
	if err != nil {
		return results, util.NewError(err, "Database error", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var temp Alerts
		err := rows.Scan(&temp.Category, &temp.TargetId, &temp.Content)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		results = append(results, temp)
	}
	return results, nil
}

func DeleteAlert(db *sql.DB, user int, category string, targetId int) error {
	stmt, err := db.Prepare(`
		DELETE
			FROM alerts
			WHERE user = ?
			AND category = ?
			AND target_id = ?;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(user, category, targetId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func createAlertContent(category string, targetId int) (string, error) {
	switch category {
		case "pending":
			return "1", nil
		case "message":
			return "2", nil
		case "accepted":
			return "3", nil
		case "removed":
			return "4", nil
	}
	return "", nil
}

func CreateAlert(db *sql.DB, user int, category string, targetId int) error {
	content, err := createAlertContent(category, targetId)
	if err != nil {return err}
	
	stmt, err := db.Prepare(`
		INSERT INTO alerts (user, category, target_id, content)
			VALUES ?,?,?;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(user, category, targetId, content)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}