package gen

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"code.google.com/p/go.crypto/bcrypt"
	"regexp"
	"encoding/hex"
	"crypto/sha256"
	"net/http"
	"net/smtp"
	"io/ioutil"
	"encoding/json"
	"crypto/tls"
	"data/util"
	"errors"
)

type unauthedUser struct {
	name string
	email string
	password string
	auth string
}

func unusedUsername(db *sql.DB, username string) bool {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
	SELECT users.name
		FROM users
		WHERE users.name = ?
		`)
	
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 19`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 27`)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return true
	}
	return false
}

func unusedEmail(db *sql.DB, email string) bool {
	stmt, err := db.Prepare(`
		SELECT users.name
			FROM users
			WHERE users.email = ?
	`)

	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 53`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(email)
	if err != nil {
		panic(err.Error() + ` ERROR IN UNUSED EMAIL`)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return true
	}
	return false
}

func CheckUserInfo(db *sql.DB, username string, email string) error {
	if(unusedEmail(db, email)){
		return errors.New("Email is in use")
	}
	if(invalidUsername(username)){
		return errors.New("Username is in an invalid format")
	}
	if(invalidEmail(email)){
		return errors.New("Email is in an invalid format")
	}
	if(unusedUsername(db, username)){
		return errors.New("Username is in use")
	}
	return nil
}

func invalidUsername(username string) bool {
	valid, err := regexp.Match("^[a-zA-ZÁÉÍÓÑÚÜáéíóñúü0-9_-]{3,20}$", []byte(username))
	if err!= nil {
		panic(err.Error() + ` Error in the regexp checking username`)
	}
	if valid {
		return false
	}
	return true
}

func invalidEmail(email string) bool {
	valid, err := regexp.Match(`\S+\@\S+\.\S`, []byte(email))
	if err!= nil {
		panic(err.Error() + ` Error in the regexp checking username`)
	}
	if valid {
		return false
	}
	return true
}

func deleteUserAuth(db *sql.DB, email string) {
	stmt, err := db.Prepare(`
		DELETE FROM unauthed_users
			WHERE email = ?
	`)

	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 19`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	res, err := stmt.Exec(email)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 50`)
	}
	_, err = res.RowsAffected()
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 54`)
	}
}

func createUserAuth(db *sql.DB, username string, password string, email string, auth string){
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		INSERT INTO unauthed_users (name, email, password, auth)
			VALUES (?, ?, ?, ?)
		`)
	defer stmt.Close()

	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 56`)
	}
	_, err = stmt.Exec(username, email, hashPassword(password), auth)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 60`)
	}
	/*
	rowCnt, err := res.RowsAffected()
	if err != nil {
		// Log the error
	}
	*/
}

func mailUserAuth(username string, toAddress string, token string) {
	from := "admin@5sur.com"
	to := toAddress
	subject := "email subject"
	body := "Boilerplate text about completing your registration process (in spanish):\nhttps://5sur.com/auth/?t=" + token 

	// Setup message (need the carriage return \r before body)
	message := "From: " + from + "\r\n"
	message += "To: " + to + "\r\n"
	message += "Subject: " + subject + "\r\n"
	message += "\r\n" + body

	// SMTP Server info
	servername := "email-smtp.us-west-2.amazonaws.com:465"
	host := "email-smtp.us-west-2.amazonaws.com"
	auth := smtp.PlainAuth("", "AKIAJ7SSYA65O5XALJVQ", "AmVDayL8URplvu+nRDaNqI46++jGqieyOJrNBYwDKN7Q", host)

	// TLS config
	tlsconfig := &tls.Config {
		ServerName: host,
	}

	conn, err := tls.Dial("tcp",servername,tlsconfig)
	if err != nil {
		panic(err.Error() + "TLS ERROR")
		return
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		panic(err.Error() + "New Client error")
		return
	}
	defer client.Quit()

	// auth
	if err = client.Auth(auth); err != nil {
		panic(err.Error() + "Auth error")
		return
	}

	// to and from
	if err = client.Mail(from); err != nil {
		panic(err.Error() + "Mail error")
		return
	}

	// Can have multiple Rcpt calls
	if err = client.Rcpt(to); err != nil {
		panic(err.Error() + "Rcpt error")
		return
	}

	// Data
	dataWriter, err := client.Data()
	if err != nil {
		panic(err.Error() + "Data error")
		return
	}
	defer dataWriter.Close()

	_, err = dataWriter.Write([]byte(message))
	if err != nil {
		panic(err.Error() + "Write error")
		return
	}
}

func UserAuth(db *sql.DB, username string, password string, email string) {
	// Create auth token
	alphaNum := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv")
	randValue := ""
	for i := 0; i < 32; i++ {
		randValue = randValue + string(alphaNum[util.RandKey(58)])
	}
	hashed := sha256.New()
	hashed.Write([]byte(randValue))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))

	deleteUserAuth(db, email)
	createUserAuth(db, username, password, email, hashedStr)

	mailUserAuth(username, email, randValue)
}

func hashPassword (password string) []byte{
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil{
		panic(err.Error() + ` THE ERROR IS ON LINE 43`)
	}
	return hashed
}

func CreateUser(db *sql.DB, token string) (string, error){
	hashed := sha256.New()
	hashed.Write([]byte(token))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))
	stmt, err := db.Prepare(`
	SELECT u.name, u.email, u.password, u.auth
		FROM unauthed_users AS u
		WHERE u.auth = ?
		`)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 78`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(hashedStr)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 86`)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()

	userInfo := unauthedUser{}

	for rows.Next() {
		err := rows.Scan(&userInfo.name, &userInfo.email, &userInfo.password, &userInfo.auth)
		if err != nil {
			panic(err.Error() + ` THE ERROR IS ON LINE 99`)
		}
	}

	if userInfo.name == "" {
		e := errors.New("Auth tokens do not match")
		return "", e
	}

	// Always run this check before creating a user (which should only be here anyway)
	if unusedUsername(db, userInfo.name) == true {
		deleteUserAuth(db, userInfo.email)
		e := errors.New("Username already taken")
		return "", e
	}

	createUser(db, userInfo)
	return userInfo.name, nil
}

func createUser(db *sql.DB, u unauthedUser) {
	stmt, err := db.Prepare(`
		INSERT INTO users (name, email, password)
			VALUES (?, ?, ?)
		`)
	defer stmt.Close()

	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 56`)
	}
	_, err = stmt.Exec(u.name, u.email, u.password)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 60`)
	}
	deleteUserAuth(db, u.email)
	/*
	rowCnt, err := res.RowsAffected()
	if err != nil {
		// Log the error
	}
	*/
}

func CheckCredentials(db *sql.DB, username string, password string) bool {
	stmt, err := db.Prepare(`
	SELECT users.password
		FROM users
		WHERE users.name = ?
		`)
	
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 78`)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		panic(err.Error() + ` THE ERROR IS ON LINE 86`)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	
	var hashedPassword []byte;

	for rows.Next() {
		err := rows.Scan(&hashedPassword)
		if err != nil {
			panic(err.Error() + ` THE ERROR IS ON LINE 99`)
		}
	}

	if hashedPassword == nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return false
	}
	return true
}

func CheckCaptcha(formValue string, userIp string) (bool, error){
	// Get super secret password from external file at some point
	resp, err := http.Get("https://www.google.com/recaptcha/api/siteverify?secret=6Lcjkf8SAAAAAMAxp-geyAYnkFwZwtkMR1uhLvjQ" + "&response="+ formValue + "&remoteip=" + userIp)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() 
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	type Captcha struct {
		Success bool
		ErrorCodes []string
	}

	var captcha Captcha
	err = json.Unmarshal(contents, &captcha)
	if err != nil {
		return false, err
	}

	return captcha.Success, nil
}