package main

import (
	"net"
	"net/http"
	"net/url"
	"5sur/gen"
	"5sur/util"
	"strconv"
	"html/template"
	"strings"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func AccountAuthHandler(w http.ResponseWriter, r *http.Request) error {
	// Query string validation
	token, err := util.ValidAuthQuery(r.URL) // Returns util.QueryFields
	if err != nil { return err }

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// Authenticate and create the user account
	_, err = gen.CreateUser(db, token)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/login", 303)
	return nil
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	// HTML generation (also does listing-specific SQL calls)
	cities, err := gen.ReturnFilter(db)
	if err != nil { return err }

	err = templates.ExecuteTemplate(w, "create.html", cities)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func CreateSubmitHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	createFormPost, err := util.ValidCreateSubmit(r)
	if err != nil { return err }

	listingId, err := gen.CreateListing(db, createFormPost.Date, userId, createFormPost.Origin, createFormPost.Destination, createFormPost.Seats, createFormPost.Fee)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/dashboard/listings?i=" + strconv.FormatInt(listingId, 10), 303)
	return nil
}

func DashListingsHandler(w http.ResponseWriter, r *http.Request) error{
	token, err := util.ValidDashQuery(r.URL)
	specificListing := false
	if err == nil {
		specificListing = true
	} else {
		token = 0
	}
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	// Check post data for if a button was clicked that directed the user here.
	if specificListing {
		err = gen.DeleteAlert(db, userId, "dropped", token)
		if err != nil { return err }
		err := gen.CheckPost(db, userId, r, token)
		if err != nil { return err }
	}

	dashListings, err := gen.GetDashListings(db, userId)
	if err != nil { return err }

	var listing gen.SpecificListing
	if specificListing {
		listing, err = gen.SpecificDashListing(db, dashListings, token)
		if err != nil { return err }
	}

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: "Dashboard",
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	body := &gen.DashListingsHTML{
		SidebarListings: dashListings,
		Listing: listing,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.DashListingsHTML
	}{
		*header,
		*body,
	}

	err = templates.ExecuteTemplate(w, "dashListings.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func DashMessagesHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	dashMessages, err := gen.GetDashMessages(db, userId)
	if err != nil { return err }

	messages := gen.MessageThread{}
	token, err := util.ValidDashQuery(r.URL) // Ignore error here
	if err == nil {
		err = gen.DeleteAlert(db, userId, "message", token)
		if err != nil { return err }
		messages, err = gen.SpecificDashMessage(db, dashMessages, token, userId)
		if err != nil { return err }
		err = gen.SetMessagesClosed(db, token, userId)
		if err != nil { return err }
		for key := range dashMessages {
			if dashMessages[key].Name == messages.Name {
				dashMessages[key].Count = 0
			}
		}
	}

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: "Dashboard",
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	body := &gen.DashMessagesHTML{
		SidebarMessages: dashMessages,
		MessageThread: messages,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.DashMessagesHTML
	}{
		*header,
		*body,
	}

	err = templates.ExecuteTemplate(w, "dashMessages.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func DashReservationsHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "",0 if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	dashReservations, err := gen.GetDashReservations(db, userId)
	if err != nil { return err }
	
	reservation := gen.Reservation{}
	token, err := util.ValidDashQuery(r.URL)
	if err == nil {
		reservation, err = gen.SpecificDashReservation(db, dashReservations, token)
		if err != nil { return err }
		err = gen.DeleteAlert(db, userId, "accepted", token)
		if err != nil { return err }
	} else {
		err = gen.DeleteAlert(db, userId, "removed", 0)
		if err != nil { return err }
		err = gen.DeleteAlert(db, userId, "deleted", 0)
		if err != nil { return err }
	}

	url, err := gen.CheckReservePost(db, userId, r, token)
	if err != nil { return err }
	if url != "" {
		http.Redirect(w, r, url, 303)
		return nil
	}
	
	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: "Dashboard",
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	body := &gen.DashReservationsHTML{
		SidebarReservations: dashReservations,
		Reservation: reservation,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.DashReservationsHTML
	}{
		*header,
		*body,
	}

	err = templates.ExecuteTemplate(w, "dashReservations.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func DeleteListingHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "",0 if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	if r.PostFormValue("d") == "" {
		listingId, err := util.ValidDashQuery(r.URL)
		if err != nil { return err }
		err = templates.ExecuteTemplate(w, "deleteListing.html", listingId)
		if err != nil {
			return util.NewError(err, "No se cargó la página", 500)
		}
		return nil
	}
	listingId, err := strconv.Atoi(r.FormValue("d"))
	if err != nil {
		return util.NewError(nil, "Viaje invalido", 400)
	}

	registeredUsers, err := gen.DeleteListing(db, userId, listingId)
	if err != nil { return err }

	for _, value := range registeredUsers {
		err = gen.CreateAlert(db, value.Id, "deleted", listingId)
		if err != nil { return err }
		err = gen.DeleteAlert(db, value.Id, "accepted", listingId)
		if err != nil { return err }
		err = gen.DeleteAlert(db, value.Id, "removed", listingId)
		if err != nil { return err }
	}

	err = gen.DeleteAlert(db, userId, "pending", listingId)
	if err != nil { return err }
	err = gen.DeleteAlert(db, userId, "dropped", listingId)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/dashboard/listings", 303)
	return nil
}

func ListingsHandler(w http.ResponseWriter, r *http.Request) error {	
	// Convert POST to GET (also does a time validation)
	if r.FormValue("Origin") != "" || r.FormValue("Destination") != "" {
		convertedDate := ""
		convertedTime := ""
		var err error
		if r.FormValue("Date") == "" {
			convertedDate, convertedTime = util.ReturnCurrentTimeString(true)
		} else if r.FormValue("Time") == "" {
			convertedDate, _, err = util.ReturnTimeString(false, r.FormValue("Date"), "00:00")
			currentDate, currentTime := util.ReturnCurrentTimeString(true)
			if currentDate == convertedDate {
				convertedTime = currentTime
			} else {
				convertedTime = "00:00"
			}
		} else {
			convertedDate, convertedTime, err = util.ReturnTimeString(false, r.FormValue("Date"), r.FormValue("Time"))
			if err != nil { return err }
		}
		http.Redirect(w, r, "https://5sur.com/l/?o=" + r.FormValue("Origin") + "&d=" + r.FormValue("Destination") + "&t=" + convertedDate + "&h=" + convertedTime, 303)
		return nil
	}

	// Query string validation
	query, err := util.ValidListingQuery(r.URL)
	if err != nil { return err }


	// Database initialization
	db, err := util.OpenDb()
	if err!=nil {
		return err
	}
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: "Viajes",
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	cities, err := gen.ReturnFilter(db)
	if err != nil { return err }

	listings, err := gen.ReturnListings(db, query.Origin, query.Destination, query.Date + " " + query.Time)
	if err != nil { return err	}

	// Convert date to be human readable
	query.Date, query.Time, err = util.ReturnTimeString(true, query.Date, query.Time)
	if err != nil { return err }

	body := &gen.ListingsHTML{
		Filter: cities,
		Listings: listings,
		Query: query,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.ListingsHTML
	}{
		*header,
		*body,
	}

	err = templates.ExecuteTemplate(w, "listings.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) error {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		return util.NewError(nil, "Falta nombre de usuario o contraseña", 400)
	}

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()


	userIp := ""
	if ipProxy := r.Header.Get("X-Real-IP"); len(ipProxy) > 0 {
		userIp = ipProxy
	} else {
		userIp, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// Check for captcha if login attempts > 2
	attempts, err := gen.CheckAttempts(db, userIp)
	if err != nil { return err }

	if attempts > 2 {
		human, err := gen.CheckCaptcha(r.FormValue("g-recaptcha-response"), userIp)
		if err != nil { return err }
		if !human {
			return util.NewError(nil, "Captcha invalido", 400)
		}
	}
	
	// User authentication
	authenticated, err := gen.CheckCredentials(db, r.FormValue("Username"), r.FormValue("Password"))
	if err != nil { return err }
	if authenticated {
		persistent := false
		if r.FormValue("Persistent") == "true" {
			persistent = true
		}
		myCookie, err := util.CreateCookie(r.FormValue("Username"), db, persistent, false) // This also stores a hashed cookie in the database
		if err != nil { return err }
		http.SetCookie(w, &myCookie)
		http.Redirect(w, r, "https://5sur.com/", 303)
		return nil
	}else {
		err = gen.UpdateLoginAttempts(db, userIp)
		if err != nil { return err }
		return util.NewError(nil, "Nombre de usuario o contraseña incorrecto", 400)
	}
	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err!=nil {
		return err
	}
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	// Create gen.InvalidateCookie
	err, expiredCookie := util.DeleteCookie(db, userId, false)
	if err != nil { return err }
	http.SetCookie(w, &expiredCookie)

	http.Redirect(w, r, "https://5sur.com/", 303)
	return nil
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) error {
	err := util.ValidRegister(r)
	if err != nil { return err }

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	err = gen.CheckUserInfo(db, r.FormValue("Username"), r.FormValue("Email"))
	if err != nil { return err }

	userIp := ""
	if ipProxy := r.Header.Get("X-Real-IP"); len(ipProxy) > 0 {
		userIp = ipProxy
	} else {
		userIp, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	human, err := gen.CheckCaptcha(r.FormValue("g-recaptcha-response"), userIp)
	if err != nil { return err }
	if !human {
		return util.NewError(nil, "Captcha invalido", 400)
	}
	
	err = gen.UserAuth(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))
	if err != nil { return err }
	Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Regístrate",
		"",
		"Email de confirmacion ha sido mandado a " + r.FormValue("Email"),
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func ReserveFormHandler(w http.ResponseWriter, r *http.Request) error {
	l, err := util.ValidReserveURL(r)
	if err != nil { return err }

	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	listing, err := gen.ReturnIndividualListing(db, l)
	if err != nil { return err }

	seats := make ([]int, 0)

	for i := 1; i <= listing.Seats ; i++ {
		seats = append(seats, i)
	}

	driver, err := gen.ReturnUserInfo(db, listing.Driver)

	reserve := &gen.ReserveHTML {
		ListingId: listing.Id,
		Driver: driver.Name,
		Seats: seats,
	}

	err = templates.ExecuteTemplate(w, "reserve.html", reserve)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) error {
	//Check POST data
	values, err := util.ValidReservePost(r)
	if err != nil { return err }

	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	err = gen.CreateReservation(db, userId, values.ListingId, values.Seats, r.FormValue("Message"))
	if err != nil { return err }

	Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Reservar",
		"Has entrado a la lista de reservaciones",
		"Atento: tu viaje no esta garantizado hasta que el conductor te acepte. Te notificaremos cuando esto suceda.",
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func RootHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	cities, err := gen.ReturnFilter(db)
	if err != nil { return err }

	listings, err := gen.ReturnAllListings(db)
	if err != nil { return err }

	body := &gen.ListingsHTML{
		Filter: cities,
		Listings: listings,
		Homepage: true,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.ListingsHTML
	}{
		*header,
		*body,
	}

	err = templates.ExecuteTemplate(w, "listings.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func UploadHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("Picture")
	if err != nil {
		return util.NewError(nil, "Foto no encontrado", 400)
	}
	defer file.Close()

	err = util.SaveImage(db, user, file, header)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/dashboard/settings", 303)
	return nil
}

func UploadFormHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	err = templates.ExecuteTemplate(w, "upload.html", "")
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func UploadDeleteFormHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	picture, err := gen.ReturnUserPicture(db, userId, "100")
	if err != nil { return err }

	body := struct {
		User int
		Picture string
	}{
		userId,
		picture,
	}

	err = templates.ExecuteTemplate(w, "deletePicture.html", body)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func UploadDeleteHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	if r.FormValue("User") != strconv.Itoa(userId) {
		return util.NewError(nil, "Foto no borrado", 400)
	}

	err = util.DeletePicture(db, user)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/dashboard/settings", 303)
	return nil
}

func UserHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	splits := strings.Split(r.URL.Path, "/")
	userInfo, err := gen.ReturnUserInfo(db, splits[2])
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: user,
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	page := struct {
		Header gen.HeaderHTML
		Body gen.User
	}{
		*header,
		userInfo,
	}

	err = templates.ExecuteTemplate(w, "user.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) error{
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if userId != 0 {
		http.Redirect(w, r, "https://5sur.com/", 303)
		return nil
	}

	userIp := ""
	if ipProxy := r.Header.Get("X-Real-IP"); len(ipProxy) > 0 {
		userIp = ipProxy
	} else {
		userIp, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	attempts, err := gen.CheckAttempts(db, userIp)
	if err != nil { return err }

	var script, captcha template.HTML
	if attempts > 2 {
		script = `<script src='https://www.google.com/recaptcha/api.js'></script>`
		captcha = `<div class="g-recaptcha" data-sitekey="6LfejAATAAAAAK1DA4l33OntwJy9LZz1GK3F2Egr"></div>`
	}
	registerData := &gen.LoginHTML {
		Title: "Ingresar",
		Script: script,
		Captcha: captcha,
	}
	err = templates.ExecuteTemplate(w, "login.html", registerData)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func RegisterFormHandler(w http.ResponseWriter, r *http.Request) error{
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	err = templates.ExecuteTemplate(w, "register.html", "")
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) error {
	recipientId, err := util.ValidMessageURL(r)
	if err != nil { return err }

	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	userInfo, err := gen.ReturnUserInfo(db, recipientId)
	if err != nil { return err }

	err = templates.ExecuteTemplate(w, "message.html", userInfo)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func SendMessageSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	recipient, message, err := util.ValidMessagePost(r)
	if err != nil { return err }

	err = gen.MessageLimit(db, userId)
	if err != nil { return err }

	err = gen.SendMessage(db, userId, recipient, message)
	if err != nil { return err }
	
	http.Redirect(w, r, "https://5sur.com/dashboard/messages?i=" + strconv.Itoa(recipient), 303)
	return nil
}

func RateHandler(w http.ResponseWriter, r *http.Request) error {
	recipientId, err := util.ValidRateURL(r)
	if err != nil { return err }

	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	userInfo, err := gen.ReturnUserInfo(db, recipientId)
	if err != nil { return err }

	err = gen.DeleteAlert(db, userId, "rate", userInfo.Id)
	if err != nil { return err }

	err = templates.ExecuteTemplate(w, "rate.html", userInfo)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func RateSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	userRate, positive, comment, public, err := util.ValidRatePost(r)
	if err != nil { return err }

	err = gen.SubmitRating(db, userId, userRate, positive, comment, public)
	if err != nil { return err }
	
		Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Dar puntaje",
		"",
		"Rating entregado!",
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func PasswordResetFormHandler(w http.ResponseWriter, r *http.Request) error {
	err := templates.ExecuteTemplate(w, "passwordReset.html", "")
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func PasswordResetHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	err = gen.ResetPassword(db, r.FormValue("Email"))
	if err != nil { return err }

	Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Restablecer contraseña",
		"",
		"Email para reestablecer contraseña ha sido mandado a " + r.FormValue("Email"),
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func PasswordChangeFormHandler(w http.ResponseWriter, r *http.Request) error {
	token, user, err := util.ValidChangePasswordQuery(r.URL)
	if err !=nil { return err }
	body := struct {
		User string
		Token string
	}{
		user,
		token,
	}
	err = templates.ExecuteTemplate(w, "passwordChange.html", body)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func PasswordChangeHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	err = util.ValidChangePasswordSubmit(r)
	if err != nil { return err }

	err = gen.ChangePassword(db, r.FormValue("User"), r.FormValue("Token"), r.FormValue("Password"))
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/login", 303)
	return nil
}

func DashSettingsHandler(w http.ResponseWriter, r *http.Request) error {	
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	user, userId, userImg, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Title: "Dashboard",
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: userImg,
	}

	page := struct {
		Header gen.HeaderHTML
	}{
		*header,
	}

	err = templates.ExecuteTemplate(w, "dashSettings.html", page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func DeleteAccountFormHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	err = templates.ExecuteTemplate(w, "deleteAccount.html", "")
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	// put this in valid
	if r.FormValue("Password") == "" || r.FormValue("Password2") == "" {
		return util.NewError(nil, "Rellena el formulario completo por favor", 400)
	}

	if r.FormValue("Password") != r.FormValue("Password2") {
		return util.NewError(nil, "No coincide la contraseña", 400)
	}

	authenticated, err := gen.CheckCredentials(db, user, r.FormValue("Password"))
	if err != nil { return err }

	if !authenticated {
		return util.NewError(nil, "Contraseña incorrecta", 400)
	}

	err = gen.DeleteAccount(db, userId)
	if err !=nil { return err }

	Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Borrar cuenta",
		"",
		"Cuenta eliminada",
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func EmailPrefHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}

	prefs, err := util.ReturnEmailPref(db, userId)
	if err != nil { return err }

	err = templates.ExecuteTemplate(w, "emailPref.html", prefs)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

func EmailPrefSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := util.OpenDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Se requiere ingreso a la cuenta", 401)
	}
	
	err = util.SetEmailPref(db, r, userId)
	if err !=nil { return err }

	Page := struct {
		Title string
		MessageTitle string
		Message string
	}{
		"Preferencias email",
		"",
		"Preferencias guardadas",
	}

	err = templates.ExecuteTemplate(w, "formSubmit.html", Page)
	if err != nil {
		return util.NewError(err, "No se cargó la página", 500)
	}
	return nil
}

type handlerWrapper func(http.ResponseWriter, *http.Request) error

func (fn handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if myErr, ok := err.(util.MyError); ok {
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}

			w.WriteHeader(myErr.StatusCode)

			ref := r.Referer()
			refUrl, _ := url.Parse(ref)
			refHost := refUrl.Host
			prevUrlText := "Volver"
			if myErr.StatusCode == 401{
				ref = "https://5sur.com/login"
				prevUrlText = "Ingresar"
			}
			if refHost != "5sur.com" {
				ref = "https://5sur.com/"
				prevUrlText = "Volver a homepage"
			}

			ErrorPage := struct {
				Code int
				Message string
				PrevUrl string
				PrevUrlText string
			}{
				myErr.StatusCode,
				myErr.Message,
				ref,
				prevUrlText,
			}

			err = templates.ExecuteTemplate(w, "error.html", ErrorPage)
			if err != nil {
				err = util.NewError(err, myErr.Error(), 500)
				util.PrintLog(err.(util.MyError))
			}
		}
	}
}

/*
type appHandlerWrapper func(http.ResponseWriter, *http.Request) error

func (fn appHandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if myErr, ok := err.(util.MyError); ok {
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}
			http.Error(w, err.Error(), myErr.StatusCode)
		}
	}
}
*/

func main() {
	util.ConfigureLog()
	http.Handle("/l/", handlerWrapper(ListingsHandler))
	http.Handle("/u/", handlerWrapper(UserHandler))
	http.Handle("/register", handlerWrapper(RegisterFormHandler))
	http.Handle("/registerSubmit", handlerWrapper(RegistrationHandler))
	http.Handle("/login", handlerWrapper(LoginFormHandler))
	http.Handle("/loginSubmit", handlerWrapper(LoginHandler))
	http.Handle("/auth/", handlerWrapper(AccountAuthHandler))
	http.Handle("/logout", handlerWrapper(LogoutHandler))
	http.Handle("/reserveSubmit", handlerWrapper(ReserveHandler))
	http.Handle("/reserve", handlerWrapper(ReserveFormHandler))
	http.Handle("/create", handlerWrapper(CreateListingHandler))
	http.Handle("/createSubmit", handlerWrapper(CreateSubmitHandler))
	http.Handle("/dashboard", handlerWrapper(DashListingsHandler))
	http.Handle("/dashboard/listings", handlerWrapper(DashListingsHandler))
	http.Handle("/dashboard/messages", handlerWrapper(DashMessagesHandler))
	http.Handle("/dashboard/reservations", handlerWrapper(DashReservationsHandler))
	http.Handle("/dashboard/listings/delete", handlerWrapper(DeleteListingHandler))
	http.Handle("/dashboard/settings", handlerWrapper(DashSettingsHandler))
	http.Handle("/uploadSubmit", handlerWrapper(UploadHandler))
	http.Handle("/upload", handlerWrapper(UploadFormHandler))
	http.Handle("/deletePicture", handlerWrapper(UploadDeleteFormHandler))
	http.Handle("/deletePictureSubmit", handlerWrapper(UploadDeleteHandler))
	http.Handle("/message", handlerWrapper(SendMessageHandler))
	http.Handle("/messageSubmit", handlerWrapper(SendMessageSubmitHandler))
	http.Handle("/rate", handlerWrapper(RateHandler))
	http.Handle("/rateSubmit", handlerWrapper(RateSubmitHandler))
	http.Handle("/passwordReset", handlerWrapper(PasswordResetFormHandler))
	http.Handle("/passwordResetSubmit", handlerWrapper(PasswordResetHandler))
	http.Handle("/passwordChange", handlerWrapper(PasswordChangeFormHandler))
	http.Handle("/passwordChangeSubmit", handlerWrapper(PasswordChangeHandler))
	http.Handle("/deleteAccount", handlerWrapper(DeleteAccountFormHandler))
	http.Handle("/deleteAccountSubmit", handlerWrapper(DeleteAccountHandler))
	http.Handle("/emailPreferences", handlerWrapper(EmailPrefHandler))
	http.Handle("/emailPrefSubmit", handlerWrapper(EmailPrefSubmitHandler))
	http.Handle("/", handlerWrapper(RootHandler))

	/*
	http.Handle("/a/logout", appHandlerWrapper(app.LogoutHandler))
	http.Handle("/a/login", appHandlerWrapper(app.LoginHandler))
	http.Handle("/a/listings", appHandlerWrapper(app.ListingsHandler))
	http.Handle("/a/listings/", appHandlerWrapper(app.ListingsHandler))
	http.Handle("/a/cities", appHandlerWrapper(app.CityHandler))
	http.Handle("/a/reserve", appHandlerWrapper(app.ReserveHandler))
	http.Handle("/a/u/", appHandlerWrapper(app.UserHandler))
	http.Handle("/a/dashboard/listings", appHandlerWrapper(app.DashListingsHandler))
	*/
	http.ListenAndServe(":8080", nil)
}