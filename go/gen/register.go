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
	"5sur/util"
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

func unusedEmail(db *sql.DB, email string) (string, error) {
	stmt, err := db.Prepare(`
		SELECT users.name
			FROM users
			WHERE users.email = ?
	`)

	if err != nil {
		return "", util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	username := ""
	err = stmt.QueryRow(email).Scan(&username)
	if err != nil {
		return "", nil
	}

	return username, nil
}

func CheckUserInfo(db *sql.DB, username string, email string) error {
	err := invalidUsername(username)
	if err != nil { return err }

	err = invalidEmail(email)
	if err != nil { return err }

	unused, err := unusedEmail(db, email)
	if err != nil { return err }

	if unused != "" { return util.NewError(nil, "Email is already in use", 400) }


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
	return util.NewError(nil, "Invalid email", 400)
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

func mailUser(toAddress string, body string) error {
	from := "admin@5sur.com"
	to := toAddress
	subject := "email subject"

	// Setup message (need the carriage return \r before body)
	message := "From: " + from + "\r\n"
	message += "To: " + to + "\r\n"
	message += "Subject: " + subject + "\r\n"
	message += "\r\n" + body

	// SMTP Server info
	user, err := ioutil.ReadFile("sesUser")
	if err != nil {
		return util.NewError(err, "Internal server error", 500)
	}
	password, err := ioutil.ReadFile("sesPassword")
	if err != nil {
		return util.NewError(err, "Internal server error", 500)
	}
	servername := "email-smtp.us-west-2.amazonaws.com:465"
	host := "email-smtp.us-west-2.amazonaws.com"
	auth := smtp.PlainAuth("", string(user[:]), string(password[:]), host)

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
		num, err := util.RandKey(58)
		if err != nil {return err}
		randValue = randValue + string(alphaNum[num])
	}
	hashed := sha256.New()
	hashed.Write([]byte(randValue))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))

	err := deleteUserAuth(db, email)
	if err != nil { return err }

	err = createUserAuth(db, username, password, email, hashedStr)
	if err != nil { return err }

	body := "Welcome to 5sur.com! Click on the following link to complete the registration process:\nhttps://5sur.com/auth/?t=" + randValue 

	err = mailUser(email, body)
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
		return "", util.NewError(err, "Creating user failed", 500)
	}
	defer stmt.Close()

	rows, err := stmt.Query(hashedStr)
	if err != nil {
		return "", util.NewError(err, "Creating user failed", 500)
	}
	defer rows.Close()

	// The last rows.Next() call will encounter an EOF error and call rows.Close()

	userInfo := unauthedUser{}

	for rows.Next() {
		err := rows.Scan(&userInfo.name, &userInfo.email, &userInfo.password, &userInfo.auth)
		if err != nil {
			return "", util.NewError(err, "Creating user failed", 500)
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
		return "", util.NewError(nil, "Username is taken", 400)
	}

	err = createUser(db, userInfo)
	if err != nil { return "", err }
	return userInfo.name, nil
}

func createUser(db *sql.DB, u unauthedUser) error {
	stmt, err := db.Prepare(`
		INSERT INTO users (name, email, password)
			VALUES (?, ?, ?)
		`)
	defer stmt.Close()

	if err != nil {
		return util.NewError(err, "Internal server error", 500)
	}
	_, err = stmt.Exec(u.name, u.email, u.password)
	if err != nil {
		return util.NewError(err, "Internal server error", 500)
	}
	deleteUserAuth(db, u.email)
	/*
	rowCnt, err := res.RowsAffected()
	if err != nil {
		// Log the error
	}
	*/
	return nil
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
	secretKey, err := ioutil.ReadFile("captchaPassword")
	if err != nil {
		return false, util.NewError(err, "Internal server error", 500)
	}
	resp, err := http.Get("https://www.google.com/recaptcha/api/siteverify?secret=" + string(secretKey[:]) + "&response="+ formValue + "&remoteip=" + userIp)
	if err != nil {
		return false, util.NewError(err, "Captcha authentication error", 500)
	}
	defer resp.Body.Close() 
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, util.NewError(err, "Captcha authentication error", 500)
	}

	type Captcha struct {
		Success bool
		ErrorCodes []string
	}

	var captcha Captcha
	err = json.Unmarshal(contents, &captcha)
	if err != nil {
		return false, util.NewError(err, "Captcha authentication error", 500)
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
	
	return nil
}

func ResetPassword(db *sql.DB, email string) error {
	if email == "" {
		return util.NewError(nil, "Email required", 400)
	}

	username, err := unusedEmail(db, email)
	if err != nil { return err }

	if username == "" { return util.NewError(nil, "Email is not registered by any user", 400) }

	alphaNum := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuv")
	randValue := ""
	for i := 0; i < 32; i++ {
		num, err := util.RandKey(58)
		if err != nil {return err}
		randValue = randValue + string(alphaNum[num])
	}
	hashed := sha256.New()
	hashed.Write([]byte(randValue))
	hashedStr := hex.EncodeToString(hashed.Sum(nil))

	err = storePasswordToken(db, email, hashedStr)
	if err != nil { return err }

	body := "Username: " + username + "\nTo reset your password, click the following link: https://5sur.com/passwordChange?t=" + randValue + "&u=" + username
	err = mailUser(email, body)

	return nil
}

func storePasswordToken(db *sql.DB, email string, token string) error {	
	stmt, err := db.Prepare(`
		INSERT INTO reset_password (email, auth)
			VALUES (?, ?)
			ON DUPLICATE KEY UPDATE
			auth = ?, created = NOW();
		`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	defer stmt.Close()

	_, err = stmt.Exec(email, token, token)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	
	return nil
}

func ChangePassword(db *sql.DB, user string, token string, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil { return err }

	hashed := sha256.New()
	hashed.Write([]byte(token))
	hashedToken := hex.EncodeToString(hashed.Sum(nil))

	stmt, err := db.Prepare(`
		UPDATE users AS u
			LEFT JOIN reset_password AS r
			ON u.email = r.email
			SET u.password = ?
			WHERE u.name = ?
			AND r.auth = ?;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	defer stmt.Close()

	_, err = stmt.Exec(hashedPassword, user, hashedToken)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}

	err = deletePasswordToken(db, hashedToken)
	if err != nil { return err }

	return nil
}

func deletePasswordToken(db *sql.DB, token string) error {	
	stmt, err := db.Prepare(`
		DELETE FROM reset_password
			WHERE auth = ?;
	`)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	defer stmt.Close()

	_, err = stmt.Exec(token)
	if err != nil {
		return util.NewError(err, "Database error", 500)
	}
	
	return nil
}