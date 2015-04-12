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
		if err != nil {
			// Log this probably
		} else {
			results = append(results, template.HTML(content))
		}
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
					A ride you were registered for has been <b>deleted</b>
				</a>
			</li>`, nil
		case "rate":
			user, err := ReturnUserInfo(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/rate?i=` + id + `">
					You can now give <b>` + user.Name + `</b> a rating based on your recent ride with them
				</a>
			</li>`, nil
	}
	return "", nil
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

	err = emailAlert(db, user, category, targetId)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func emailAlert(db *sql.DB, user int, category string, targetId int) error {
	send, err := emailPref(db, user, category)
	if err != nil { return util.NewError(err, "Database error", 500) }

	toAddress, err := returnUserEmail(db, user)
	if err != nil { return util.NewError(err, "Database error", 500) }

	// If email pref set to not email for that category, return nil and send no email
	if !send {
		return nil
	}

	subject := ""
	text := ""
	buttonText := ""
	buttonLink := ""

	id := strconv.Itoa(targetId)
	switch category {
		case "pending": // The target id is the listing id
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - New pending users on your listing"
			text = `You have new pending users on listing ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "View Listing"
			buttonLink = "https://5sur.com/dashboard/listings?i=" + id
		case "dropped":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - Someone has dropped from one of your listings"
			text = `Someone has dropped from ride ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "View Listing"
			buttonLink = "https://5sur.com/dashboard/listings?i=" + id
		case "message": // The target id is the user id
			message, err := returnAlertMessage(db, user, targetId)
			if err != nil {return err}
			subject = "5sur - New message"
			text = `You have a new message from ` + message.Name + `.`
			buttonText = "View Message"
			buttonLink = "https://5sur.com/dashboard/messages?i=" + id
		case "accepted":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - You have been accepted onto a ride"
			text = `You have been Accepted onto ride ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "View Reservation"
			buttonLink = "https://5sur.com/dashboard/reservations?i=" + id
		case "removed":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - You have been removed from a ride you were registered for"
			text = `You have been removed from ride ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "View Reservations"
			buttonLink = "https://5sur.com/dashboard/reservations"
		case "deleted":
			subject = "5sur - Listing deleted"
			text = `A ride you were registered for has been <b>deleted</b> by the driver.`
			buttonText = "View Reservations"
			buttonLink = "https://5sur.com/dashboard/reservations"
	}

	body := returnBody(text, buttonText, buttonLink)

	err = util.SendEmail(toAddress, subject, body)
	if err != nil { return util.NewError(err, "Database error", 500) }

	return nil
}

// Return true if you want an email, false otherwise
func emailPref(db *sql.DB, user int, category string) (bool, error) {
	prefs, err := util.ReturnEmailPref(db, user)
	if err != nil { return false, err }

	switch category {
		case "pending":
			return prefs.Pending, nil
		case "dropped":
			return prefs.Dropped, nil
		case "message":
			return prefs.Message, nil
		case "accepted":
			return prefs.Accepted, nil
		case "removed":
			return prefs.Removed, nil
		case "deleted":
			return prefs.Deleted, nil
	}

	return false, nil
}

func returnUserEmail(db *sql.DB, user int) (string, error) {
	stmt, err := db.Prepare(`
		SELECT u.email
			FROM users AS u
			WHERE u.id = ?;
		`)
	
	if err != nil {
		return "", util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	email := ""
	err = stmt.QueryRow(user).Scan(&email)
	if err != nil {
		return "", util.NewError(err, "Internal Server Error", 500)
	}

	return email, nil
}

func returnBody(text string, buttonText string, buttonLink string) string {
	return `<!doctype html>
<html>
<head>
	
</head>
<body>
<table width="100%" bgcolor="#A5DCF1">
	<tr height="35px"></tr>
	<tr width="100%">
			<td width="100%">
	<center>
	<table cellspacing="0" cellpadding="0" bgcolor="#FFFFFF" style="-webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; box-shadow: 0px 1px 2px #98A8AD;">
		<tr height="50px"></tr>
		<tr>
			<td width="25px"></td>
			<td width="100px"></td>
			<td width="250px" height="89px" align="center">
				<a href="https://5sur.com/" style="text-decoration: none; width:100%; height:100%; display:block">
					<img src="https://5sur.com/graphics/logo2.png" alt="5sur">
				</a></td>
			<td width="100px"></td>
			<td width="25px"></td>
		</tr>
		<tr height="50px"></tr>
		<tr>
			<td width="25px"></td>
			<td width="350px" colspan="3" align="center" style="color:#414243; font-size:16px;">` + text + `</td>
			<td width="25px"></td>
		</tr>
		<tr height="50px"></tr>
		<tr>
			<td width="25px"></td>
			<td width="100px"></td>
			<td align="center" width="250px" height="50px" bgcolor="#4DBBE6" style="-webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; color: #ffffff; display: block;">
			<a href="` + buttonLink + `" style="font-size:16px; font-weight: bold; font-family: Helvetica, Arial, sans-serif; text-decoration: none; line-height:50px; width:100%; display:inline-block"><span style="color: #FFFFFF">` + buttonText + `</span></a>

			</td>
			<td width="100px"></td>
			<td width="25px"></td>
		</tr>
		<tr height="50px"></tr>
		<tr height="1px">
			<td width="25px"></td>
			<td colspan="3" bgcolor="#cccccc"></td>
			<td width="25px"></td>
		</tr>
		<tr height="15px"></tr>
		<tr height="1px">
			<td width="25px"></td>
			<td colspan="3" align="center" style="font-size:12px; color=#666666;">Unsuscribe from emails by changing your <a href="https://5sur.com/emailPreferences" style="text-decoration: none; color=#4DBBE6">email preferences</a></td>
			<td width="25px"></td>
		</tr>
		<tr height="35px"></tr>
	</table> 
	</center>
	</td>
	</tr>
	<tr height="35px"></tr>
</table>
	
</body>
</html>`
}