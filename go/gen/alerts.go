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
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(user)
	if err != nil {
		return results, util.NewError(err, "Error de la base de datos", 500)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var targetId int
		err := rows.Scan(&category, &targetId)
		if err != nil {
			return results, util.NewError(err, "Error de la base de datos", 500)
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
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user, category, targetId)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
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
					Nuevos usuarios <b>pendientes</b> en el viaje ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>`, nil
		case "dropped":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/listings?i=` + id + `">
					Alguien <b>abortó</b> el viaje ` + listing.Origin + ` > ` + listing.Destination + `
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
					Viaje <b>aceptada</b> ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>
			`, nil
		case "removed":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/dashboard/reservations">
					Has sido <b>eliminado</b> del viaje ` + listing.Origin + ` > ` + listing.Destination + `
				</a>
			</li>
			`, nil
		case "deleted":
			return `
			<li>
				<a href="https://5sur.com/dashboard/reservations">
					El viaje donde te registraste ha sido <b>borrado</b>
				</a>
			</li>`, nil
		case "rate":
			user, err := ReturnUserInfo(db, targetId)
			if err != nil {return "", err}
			return `
			<li>
				<a href="https://5sur.com/rate?i=` + id + `">
					Ahora puedes dar un rating a <b>` + user.Name + `</b> por un viaje recién compartido
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
		return AlertMessage{}, util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()
	customPicture := false
	err = stmt.QueryRow(recipient, sender).Scan(&result.Name, &customPicture, &result.Message)
	if err != nil {
		return AlertMessage{}, util.NewError(nil, "Mensaje no existe", 400)
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
		return util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user, category, targetId, user, category, targetId)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}

	err = emailAlert(db, user, category, targetId)
	if err != nil {
		return util.NewError(err, "Error de la base de datos", 500)
	}
	return nil
}

func emailAlert(db *sql.DB, user int, category string, targetId int) error {
	send, err := emailPref(db, user, category)
	if err != nil { return util.NewError(err, "Error de la base de datos", 500) }

	toAddress, err := returnUserEmail(db, user)
	if err != nil { return util.NewError(err, "Error de la base de datos", 500) }

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
			subject = "5sur - Nuevos usuarios pendientes"
			text = `Nuevos usuarios pendientes en el viaje ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "Ver listing"
			buttonLink = "https://5sur.com/dashboard/listings?i=" + id
		case "dropped":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - Alguien se ha retirado de tu viaje"
			text = `Alguien se ha retirado de tu viaje ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "Ver listing"
			buttonLink = "https://5sur.com/dashboard/listings?i=" + id
		case "message": // The target id is the user id
			message, err := returnAlertMessage(db, user, targetId)
			if err != nil {return err}
			subject = "5sur - Mensaje nuevo"
			text = `Tienes un mensaje nuevo de ` + message.Name + `.`
			buttonText = "Ver mensaje"
			buttonLink = "https://5sur.com/dashboard/messages?i=" + id
		case "accepted":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - Has sido aceptado por un viaje"
			text = `Has sido aceptado por viajee ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "Ver reserva"
			buttonLink = "https://5sur.com/dashboard/reservations?i=" + id
		case "removed":
			listing, err := ReturnIndividualListing(db, targetId)
			if err != nil {return err}
			subject = "5sur - You have been removed from a ride you were registered for"
			text = `You have been removed from ride ` + listing.Origin + ` > ` + listing.Destination + `.`
			buttonText = "Ver reservas"
			buttonLink = "https://5sur.com/dashboard/reservations"
		case "deleted":
			subject = "5sur - Listing eliminado"
			text = `Un viaje por lo cual estabas registrado ha sido eliminado por el conductor.`
			buttonText = "Ver reservas"
			buttonLink = "https://5sur.com/dashboard/reservations"
	}

	body := util.EmailTemplate(text, buttonText, buttonLink)

	err = util.SendEmail(toAddress, subject, body)
	if err != nil { return err }

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
		return "", util.NewError(err, "Error de la base de datos", 500)
	}
	defer stmt.Close()

	email := ""
	err = stmt.QueryRow(user).Scan(&email)
	if err != nil {
		return "", util.NewError(err, "Error de servidor", 500)
	}

	return email, nil
}