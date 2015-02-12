package gen

import (
	"5sur/util"
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
		SELECT category, target_id
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
		var category string
		var targetId int
		err := rows.Scan(&category, &targetId)
		if err != nil {
			return results, util.NewError(err, "Database error", 500)
		}
		content, err := createAlertContent(db, user, category, targetId)
		if err != nil {return results, err}
		results = append(results, template.HTML(content))
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
					<img src="` + message.Picture + `" alt="user">
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
			user, err := ReturnUserInfo(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/rate?i=` + id + `">
					You can now give <b>` + user.Name + `</b> a rating based on your recent ride with them.
				</a>
			</li>`, nil
	}
	return "", nil
}

func CreateAlert(db *sql.DB, user int, category string, targetId int) error {
	stmt, err := db.Prepare(`
		INSERT INTO alerts (user, category, target_id)
			SELECT ? AS user, ? AS category, ? AS target_id FROM dual
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

	_, err = stmt.Exec(user, category, targetId, user, category, targetId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func returnAlertMessage(db *sql.DB, recipient int, sender int) (AlertMessage, error) {
	result := AlertMessage{}
	stmt, err := db.Prepare(`
		SELECT u.name, u.custom_picture, m.message
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
	customPicture := false
	err = stmt.QueryRow(recipient, sender).Scan(&result.Name, &customPicture, &result.Message)
	if err != nil {
		return AlertMessage{}, util.NewError(nil, "Message does not exist", 400)
	}

	if customPicture {
		result.Picture = "https://5sur.com/images/" + result.Name + "_35.png"
	} else {
		result.Picture = "https://5sur.com/default_35.png"
	}
	return result, nil
}