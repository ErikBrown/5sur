package gen

import (
	"data/util"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"html/template"
)

type AlertMessage struct {
	Name string
	Picture string
	Message string
}

func GetAlerts(db *sql.DB, user int) ([]template.HTML, error) {
	var results []template.HTML
	stmt, err := db.Prepare(`
		SELECT content
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
		var temp string
		err := rows.Scan(&temp)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		results = append(results, template.HTML(temp))
	}
	return results, nil
}

func DeleteAlert(db *sql.DB, user int, category string, targetId int) error {
	stmt, err := db.Prepare(`
		DELETE
			FROM alerts 
			WHERE user = ?
			AND category = ?
			AND (
					(
					category = "removed" 
					OR category = "deleted"
					)
					OR target_id = ?
				);
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

func createAlertContent(db *sql.DB, user int, category string, targetId int) (string, error) {
	id := strconv.Itoa(targetId)

	switch category {
		case "pending": // The target id is the listing id
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/listings?i=` + id + `">
					<b>New pending users</b> on ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>`, nil
		case "dropped":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/listings?i=` + id + `">
					Someone has <b>dropped</b> from ride ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>
			`, nil
		case "message": // The target id is the user id
			message, err := returnAlertMessage(db, user, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/messages?i=` + id + `">
					<img src="https://5sur.com/` + message.Picture + `" alt="user">
					<b>` + message.Name + `</b><p>` + message.Message + `</p>
				</a>
			</li>`, nil
		case "accepted":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/reservations?i=` + id + `">
					<b>Accepted</b> onto ride ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>
			`, nil
		case "removed":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/reservations">
					You have been <b>removed</b> from ride ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>
			`, nil
		case "deleted":
			return `
			<li>
				<a href="https://5sur.com/dashboard/reservations">
					A ride you were registered for has been <b>deleted</b>.
				</a>
			</li>`, nil
		case "rate":
			return `
			<li>
				<a href="https://5sur.com/user/` + targetId + `/rate">
					You can now give <b>`targetId`</b> a rating based on your recent ride with them.
				</a>
			</li>`, nil
	}
	return "", nil
}

func CreateAlert(db *sql.DB, user int, category string, targetId int) error {
	content, err := createAlertContent(db, user, category, targetId)
	if err != nil {return err}
	
	stmt, err := db.Prepare(`
		INSERT INTO alerts (user, category, target_id, content)
			SELECT ? AS user, ? AS category, ? AS target_id, ? AS content FROM dual
				WHERE NOT EXISTS(
					SELECT user
						FROM alerts
						WHERE user = ?
						AND category = ?
						AND target_id = ?
				) LIMIT 1;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	_, err = stmt.Exec(user, category, targetId, content, user, category, targetId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func returnAlertMessage(db *sql.DB, recipient int, sender int) (AlertMessage, error) {
	result := AlertMessage{}
	stmt, err := db.Prepare(`
		SELECT u.name, u.picture, m.message
			FROM messages AS m
			JOIN users AS u ON m.sender = u.id
			WHERE m.receiver = ?
				AND m.sender = ?
			ORDER BY m.date DESC
			LIMIT 1;
		`)
	
	if err != nil {
		return AlertMessage{}, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	err = stmt.QueryRow(recipient, sender).Scan(&result.Name, &result.Picture, &result.Message)
	if err != nil {
		return AlertMessage{}, util.NewError(nil, "Message does not exist", 400)
	}
	return result, nil
}