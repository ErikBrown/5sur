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
)

type unauthedUser struct {
	name string
	email string
	password string
	auth string
}

func unusedUsername(db *sql.DB, username string) (bool, error) {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
	SELECT users.name
		FROM users
		WHERE users.name = ?
		`)
	
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return false, nil
	}
	return true, nil
}

func unusedEmail(db *sql.DB, email string) error {
	stmt, err := db.Prepare(`
		SELECT users.name
			FROM users
			WHERE users.email = ?
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(email)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	for rows.Next() {
		return util.NewError(nil, "Email is already in use", 400)
	}
	return nil
}

func CheckUserInfo(db *sql.DB, username string, email string) error {
	err := unusedEmail(db, email)
	if err != nil { return err }

	err = invalidUsername(username)
	if err != nil { return err }

	err = invalidEmail(email)
	if err != nil { return err }

	uniqueUsername, err := unusedUsername(db, username)
	if err != nil { return err }

	if !uniqueUsername {
		return util.NewError(nil, "Username is taken", 400)
	}

	return nil
}

func invalidUsername(username string) error {
	valid, err := regexp.Match("^[a-zA-ZÁÉÍÓÑÚÜáéíóñúü0-9_-]{3,20}$", []byte(username))
	if err!= nil {
		return util.NewError(err, "Internal server error", 500)
	}
	if valid {
		return nil
	}
	return util.NewError(nil, "Invalid username", 400)
}

func invalidEmail(email string) error {
	valid, err := regexp.Match(`\S+\@\S+\.\S`, []byte(email))
	if err!= nil {
		return util.NewError(err, "Internal server error", 500)
	}
	if valid {
		return nil
	}
	return util.NewError(nil, "Invalid username", 400)
}

func deleteUserAuth(db *sql.DB, email string) error {
	stmt, err := db.Prepare(`
		DELETE FROM unauthed_users
			WHERE email = ?
	`)

	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	res, err := stmt.Exec(email)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	_, err = res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	return nil
}

func createUserAuth(db *sql.DB, username string, password string, email string, auth string) error {
	// Always prepare queries to be used multiple times. The parameter placehold is ?
	stmt, err := db.Prepare(`
		INSERT INTO unauthed_users (name, email, password, auth)
			VALUES (?, ?, ?, ?)
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(username, email, hashedPassword, auth)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	
	_, err = res.RowsAffected()
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	return nil
}

func mailUserAuth(username string, toAddress string, token string) error {
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
		return util.NewError(err, "Email authentication error", 500)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}
	defer client.Quit()

	// auth
	if err = client.Auth(auth); err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}

	// to and from
	if err = client.Mail(from); err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}

	// Can have multiple Rcpt calls
	if err = client.Rcpt(to); err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}

	// Data
	dataWriter, err := client.Data()
	if err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}
	defer dataWriter.Close()

	_, err = dataWriter.Write([]byte(message))
	if err != nil {
		return util.NewError(err, "Email authentication error", 500)
	}
	return nil
}

func UserAuth(db *sql.DB, username string, password string, email string) error {
	// Create auth token
	alphaNum := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv")
	randValue := ""
	for i := 0; i < 32; i++ {
		randValue = randValue + string(alphaNum[util.RandKey(58)])
	}
	hashed := sha256.New()
	hashed.Write([]byte(randValue))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))

	err := deleteUserAuth(db, email)
	if err != nil { return err }

	err = createUserAuth(db, username, password, email, hashedStr)
	if err != nil { return err }

	err = mailUserAuth(username, email, randValue)
	if err != nil { return err }

	return nil
}

func hashPassword(password string) ([]byte, error){
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil{
		return hashed, util.NewError(err, "Internal server error", 500)
	}
	return hashed, nil
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
		return "", util.NewError(err, "Creating user failed, please try again later", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(hashedStr)
	if err != nil {
		return "", util.NewError(err, "Creating user failed, please try again later", 500)
	}
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()

	userInfo := unauthedUser{}

	for rows.Next() {
		err := rows.Scan(&userInfo.name, &userInfo.email, &userInfo.password, &userInfo.auth)
		if err != nil {
			return "", util.NewError(err, "Creating user failed, please try again later", 500)
		}
	}

	if userInfo.name == "" {
		return "", util.NewError(nil, "Authentication failed", 400)
	}

	// Always run this check before creating a user (which should only be here anyway)
	uniqueUsername, err := unusedUsername(db, userInfo.name)
	if err != nil {
		return "", err
	}
	if !uniqueUsername {
		deleteUserAuth(db, userInfo.email)
		return "", util.NewError(nil, "Username already taken", 400)
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

func CheckCredentials(db *sql.DB, username string, password string) (bool, error) {
	stmt, err := db.Prepare(`
	SELECT users.password
		FROM users
		WHERE users.name = ?
		`)
	
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	// db.Query() prepares, executes, and closes a prepared statement - three round
	// trips to the databse. Call it infrequently as possible; use efficient SQL statments
	rows, err := stmt.Query(username)
	if err != nil {
		return false, util.NewError(err, "Database error", 500)
	}
	// Always defer rows.Close(), even if you explicitly Close it at the end of the
	// loop. The connection will have the chance to remain open otherwise.
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()
	
	var hashedPassword []byte;

	for rows.Next() {
		err := rows.Scan(&hashedPassword)
		if err != nil {
			return false, util.NewError(err, "Database error", 500)
		}
	}

	if hashedPassword == nil {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func CheckCaptcha(formValue string, userIp string) (bool, error){
	// Get super secret password from external file at some point
	resp, err := http.Get("https://www.google.com/recaptcha/api/siteverify?secret=6Lcjkf8SAAAAAMAxp-geyAYnkFwZwtkMR1uhLvjQ" + "&response="+ formValue + "&remoteip=" + userIp)
	if err != nil {
		return false, util.NewError(err, "Verification error. Please try again later.", 500)
	}
	defer resp.Body.Close() 
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, util.NewError(err, "Verification error. Please try again later.", 500)
	}

	type Captcha struct {
		Success bool
		ErrorCodes []string
	}

	var captcha Captcha
	err = json.Unmarshal(contents, &captcha)
	if err != nil {
		return false, util.NewError(err, "Verification error. Please try again later.", 500)
	}
	return captcha.Success, nil
}

func CheckAttempts(db *sql.DB, ip string) (int, error) {
	stmt, err := db.Prepare(`
		SELECT attempts
			FROM login_attempts
			WHERE ip = ?
	`)
	
	if err != nil {
		return 0, util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	attempts := 0
	err = stmt.QueryRow(ip).Scan(&attempts)
	if err != nil {
		// IP hasn't attempted to login yet
	}
	return attempts, nil
}

func UpdateLoginAttempts(db *sql.DB, ip string) error {
	stmt, err := db.Prepare(`
		INSERT INTO login_attempts (ip, attempts)
			VALUES (?, 1)
			ON DUPLICATE KEY UPDATE
			attempts = attempts + 1;
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	defer stmt.Close()

	_, err = stmt.Exec(ip)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	
	/*
	rowCnt, err := res.RowsAffected()
	if err != nil {
		// Log the error
	}
	*/
	return nil
}