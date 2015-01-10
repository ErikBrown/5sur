package main

import (
	"fmt"
	"net"
	"net/http"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"data/gen"
	"data/util"
	"strconv"
	"html/template"
	"os"
	"image"
	"image/png"
	_ "image/jpeg"
	_ "image/gif"
	// "log"
)

var templates = template.Must(template.ParseFiles("templates/login.html","templates/dashMessages.html","templates/message.html"))

func openDb() (*sql.DB, error) {
	db, err := sql.Open("mysql", "gary:butthole@/rideshare")
	if err != nil {
		return db, util.NewError(err, "Database connection failed", 500)
	}
	err = db.Ping()
	if err != nil {
		return db, util.NewError(err, "Database connection failed", 500)
	}
	return db, nil
}

func AccountAuthHandler(w http.ResponseWriter, r *http.Request) error {
	// Query string validation
	token, err := util.ValidAuthQuery(r.URL) // Returns util.QueryFields
	if err != nil { return err }

	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// Authenticate and create the user account
	user, err := gen.CreateUser(db, token)
	if err != nil { return err }

	fmt.Fprint(w, user + ", your accout is activated!")
	return nil
}

func AppListingsHandler(w http.ResponseWriter, r *http.Request) error {
	query, err := util.ValidListingQuery(r.URL) // Returns util.QueryFields
	if err != nil { return err }

	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	listings, err := gen.ReturnListings(db, query.Origin, query.Destination, query.Date + " " + query.Time)
	if err != nil { return err }
	jsonListings, err := json.MarshalIndent(listings, "", "    ")
	if err != nil {
		return util.NewError(nil, "Json conversion failed", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonListings))
	return nil
}

func CreateListingHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	// HTML generation (also does listing-specific SQL calls)
	createListingPage, err := gen.CreateListingPage(db, user)
	if err != nil { return err }

	fmt.Fprint(w, createListingPage)
	return nil
}

func CreateSubmitHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	createFormPost, err := util.ValidCreateSubmit(r)
	if err != nil { return err }

	err = gen.CreateListing(db, createFormPost.Date, userId, createFormPost.Origin, createFormPost.Destination, createFormPost.Seats, createFormPost.Fee)
	if err != nil { return err }

	fmt.Fprint(w, "listing created!")
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
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	// Check post data for if a button was clicked that directed the user here.
	if specificListing {
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

	// HTML generation
	dashListingsPage, err := gen.DashListingsPage(dashListings, listing, user);
	if err != nil { return err }

	fmt.Fprint(w, dashListingsPage)
	return nil
}

func DashMessagesHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}
	dashMessages, err := gen.GetDashMessages(db, userId)
	if err != nil { return err }

	messages := gen.MessageThread{}
	token, err := util.ValidDashQuery(r.URL) // Ignore error here
	if err == nil {
		messages, err = gen.SpecificDashMessage(db, dashMessages, token, userId)
		if err != nil { return err }
		err = gen.SetMessagesClosed(db, token, userId)
		if err != nil { return err }
		err = gen.DeleteAlert(db, userId, "message", token)
		if err != nil { return err }
		for key := range dashMessages {
			if dashMessages[key].Name == messages.Name {
				dashMessages[key].Count = 0
			}
		}
	}

	// HTML generation
	/*
	dashMessagesPage, err := gen.DashMessagesPage(dashMessages, message, user);
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, dashMessagesPage)
	*/

	alerts, err := gen.GetAlerts(db, userId)
	if err != nil { return err }

	header := &gen.HeaderHTML {
		Username: user,
		Alerts: len(alerts),
		AlertText: alerts,
		UserImage: "https://5sur.com/default.png",
	}

	body := &gen.DashMessagesHTML{
		Title: "Dashboard",
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
		return util.NewError(err, "Failed to load page", 500)
	}
	return nil
}

func DashReservationsHandler(w http.ResponseWriter, r *http.Request) error{
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckCookie(r, db) // return "",0 if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Login required", 401)
	}

	dashReservations, err := gen.GetDashReservations(db, userId)
	if err != nil { return err }
	
	reservation := gen.Reservation{}
	token, err := util.ValidDashQuery(r.URL)
	if err == nil {
		reservation, err = gen.SpecificDashReservation(db, dashReservations, token)
		if err != nil { return err }
	}

	url, err := gen.CheckReservePost(db, userId, r, token)
	if err != nil { return err }
	if url != "" {
		http.Redirect(w, r, url, 303)
		return nil
	}
	// HTML generation
	dashReservationsPage, err := gen.DashReservationsPage(dashReservations, reservation, user)
	if err != nil { return err }
	fmt.Fprint(w, dashReservationsPage)
	return nil
}

func DeleteListingHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, err := util.CheckCookie(r, db) // return "",0 if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Login required", 401)
	}

	if r.PostFormValue("d") == "" {
		listingId, err := util.ValidDashQuery(r.URL)
		if err != nil { return err }
		deleteForm := gen.DeleteForm(listingId)
		fmt.Fprint(w, deleteForm)
		return nil
	}
	listingId, err := strconv.Atoi(r.FormValue("d"))
	if err != nil {
		return util.NewError(nil, "Invalid listing", 400)
	}

	err = gen.DeleteListing(db, userId, listingId)
	if err != nil { return err }

	http.Redirect(w, r, "https://5sur.com/dashboard/listings", 303)
	return nil
}

func ListingsHandler(w http.ResponseWriter, r *http.Request) error {	
	// Convert POST to GET (also does a time validation)
	if r.FormValue("Origin") != "" && r.FormValue("Destination") != "" {
		convertedDate := ""
		convertedTime := ""
		var err error
		if r.FormValue("Date") == "" {
			convertedDate, convertedTime = util.ReturnCurrentTimeString()
		} else if r.FormValue("Time") == "" {
			convertedDate, _, err = util.ReturnTimeString(false, r.FormValue("Date"), "00:00")
			currentDate, currentTime := util.ReturnCurrentTimeString()
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
	if err != nil {
		return err
	}

	// Database initialization
	db, err := openDb()
	if err!=nil {
		return err
	}
	defer db.Close()

	// User authentication
	user, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	// HTML generation (also does listing-specific SQL calls)
	listPage, err := gen.ListingsPage(db, query, user);
	if err != nil {
		return err
	}

	fmt.Fprint(w, listPage)
	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request) error {
	// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" {
		return util.NewError(nil, "Missing username or password", 400)
	}

	// Database initialization
	db, err := openDb()
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
			return util.NewError(nil, "Incorrect Captcha", 400)
		}
	}
	
	// User authentication
	authenticated, err := gen.CheckCredentials(db, r.FormValue("Username"), r.FormValue("Password"))
	if err != nil { return err }
	if authenticated {
		myCookie, err := util.CreateCookie(r.FormValue("Username"), db) // This also stores a hashed cookie in the database
		if err != nil { return err }
		http.SetCookie(w, &myCookie)
		http.Redirect(w, r, "https://5sur.com/", 303)
		return nil
	}else {
		err = gen.UpdateLoginAttempts(db, userIp)
		if err != nil { return err }
		return util.NewError(nil, "Your username or password was incorrect", 400)
	}
	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	// Create gen.InvalidateCookie
	expiredCookie := util.DeleteCookie()
	http.SetCookie(w, &expiredCookie)

	fmt.Fprint(w, "you SHOULD be logged out")
	return nil
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) error {
	err := util.ValidRegister(r)
	if err != nil { return err }

	// Database initialization
	db, err := openDb()
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
		return util.NewError(nil, "Incorrect Captcha", 400)
	}
	
	err = gen.UserAuth(db, r.FormValue("Username"), r.FormValue("Password"), r.FormValue("Email"))
	if err != nil { return err }

	fmt.Fprint(w, "Confirmation email has been sent to " + r.FormValue("Email"))
	return nil
}

func ReserveFormHandler(w http.ResponseWriter, r *http.Request) error {
	l, err := util.ValidReserveURL(r)
	if err != nil { return err }

	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}
	listing, err := gen.ReturnIndividualListing(db, l)
	if err != nil { return err }

	reservePage := gen.CreateReserveFormPage(listing, user)
	fmt.Fprint(w, reservePage)
	return nil
}

func ReserveHandler(w http.ResponseWriter, r *http.Request) error {
	//Check POST data
	values, err := util.ValidRegisterPost(r)
	if err != nil { return err }

	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}

	err = gen.CreateReservation(db, userId, values.ListingId, values.Seats, r.FormValue("Message"))
	if err != nil { return err }

	reservePage := gen.CreateReservePage(values.ListingId, values.Seats, user, r.FormValue("Message"))

	fmt.Fprint(w, reservePage)
	return nil
}

func RootHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	// HTML generation (also does listing-specific SQL calls)
	homePage, err := gen.HomePage(db, user);
	if err != nil { return err }

	fmt.Fprint(w, homePage)
	return nil
}

func UploadHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	user, _, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }
	if user == "" {
		return util.NewError(nil, "Login required", 401)
	}
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("Picture")
	if err != nil {
		return util.NewError(nil, "No picture found", 400)
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return util.NewError(nil, "Invalid image format", 400)
	}

	// This should be userId rather than uploadedFile.png.
	// TODO
	toimg, _ := os.Create("/var/www/html/images/uploadedFile.png")
	defer toimg.Close()
	err = png.Encode(toimg, image)	
	if err != nil {
		return util.NewError(err, "Image cannot be used", 500)
	}

	bounds := image.Bounds()
	ratio:= float64(bounds.Dx())/float64(bounds.Dy())
	if ratio < .8 || ratio > 1.2 {
		return util.NewError(nil, "Invalid image dimensions", 400)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
	return nil
}

func UserHandler(w http.ResponseWriter, r *http.Request) error {
	// Database initialization
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	user, err := gen.ReturnUserInfo(db, r.URL.Path[3:])
	if err != nil { return err }

	formatted, err := json.MarshalIndent(user, "", "    ")
	if err != nil {
		return util.NewError(err, "Json conversion failed", 500)
	}
	fmt.Fprint(w, string(formatted))
	return nil
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) error{
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

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
		captcha = `<div class="g-recaptcha" data-sitekey="6Lcjkf8SAAAAAE242oMsYj9akkUm69jfYIlSBOLF"></div>`
	}
	registerData := &gen.LoginHTML{
		Title: "Login",
		Script: script,
		Captcha: captcha,
	}
	err = templates.ExecuteTemplate(w, "login.html", registerData)
	if err != nil {
		return util.NewError(err, "Failed to load page", 500)
	}
	return nil
}

func AppCityHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()
	cities, err := gen.ReturnFilter(db)
	if err != nil { return err }

	jsonCities, err := json.MarshalIndent(cities, "", "    ")
	if err != nil {
		return util.NewError(err, "Json conversion failed", 500)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonCities))
	return nil
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) error {
	recipientId, err := util.ValidMessageURL(r)
	if err != nil { return err }

	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Login required", 401)
	}
	userInfo, err := gen.ReturnUserInfo(db, recipientId)
	if err != nil { return err }

	err = templates.ExecuteTemplate(w, "message.html", userInfo)
	if err != nil {
		return util.NewError(err, "Failed to load page", 500)
	}
	return nil
}

func SendMessageSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	db, err := openDb()
	if err != nil { return err }
	defer db.Close()

	// User authentication
	_, userId, err := util.CheckCookie(r, db) // return "" if not logged in
	if err != nil { return err }

	if userId == 0 {
		return util.NewError(nil, "Login required", 401)
	}

	recipient, message, err := util.ValidMessagePost(r)
	if err != nil { return err }

	err = gen.SendMessage(db, userId, recipient, message)
	if err != nil { return err }
	
	fmt.Fprint(w, "Message submitted successfully!:" + "\n\n" + r.FormValue("Recipient") + "\n\n" + r.FormValue("Message"))
	return nil
}

type handlerWrapper func(http.ResponseWriter, *http.Request) error

func (fn handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		if myErr, ok := err.(util.MyError); ok {
			http.Error(w, myErr.Message, myErr.Code)
			if myErr.LogError != nil {
				util.PrintLog(myErr)
			}
		}
	}
}

func main() {
	util.ConfigureLog()
	http.Handle("/l/", handlerWrapper(ListingsHandler))
	http.Handle("/u/", handlerWrapper(UserHandler))
	http.Handle("/a/listings", handlerWrapper(AppListingsHandler))
	http.Handle("/a/listings/", handlerWrapper(AppListingsHandler))
	http.Handle("/a/cities", handlerWrapper(AppCityHandler))
	http.Handle("/login", handlerWrapper(LoginHandler))
	http.Handle("/register", handlerWrapper(RegistrationHandler))
	http.Handle("/loginForm", handlerWrapper(LoginFormHandler))
	http.Handle("/loginForm/", handlerWrapper(LoginFormHandler))
	http.Handle("/auth/", handlerWrapper(AccountAuthHandler))
	http.Handle("/logout", handlerWrapper(LogoutHandler))
	http.Handle("/reserveSubmit", handlerWrapper(ReserveHandler))
	http.Handle("/reserve", handlerWrapper(ReserveFormHandler))
	http.Handle("/create", handlerWrapper(CreateListingHandler))
	http.Handle("/createSubmit", handlerWrapper(CreateSubmitHandler))
	http.Handle("/dashboard/listings", handlerWrapper(DashListingsHandler))
	http.Handle("/dashboard/messages", handlerWrapper(DashMessagesHandler))
	http.Handle("/dashboard/reservations", handlerWrapper(DashReservationsHandler))
	http.Handle("/dashboard/listings/delete", handlerWrapper(DeleteListingHandler))
	http.Handle("/upload", handlerWrapper(UploadHandler))
	http.Handle("/message", handlerWrapper(SendMessageHandler))
	http.Handle("/messageSubmit", handlerWrapper(SendMessageSubmitHandler))
	http.Handle("/", handlerWrapper(RootHandler))
	http.ListenAndServe(":8080", nil)
}