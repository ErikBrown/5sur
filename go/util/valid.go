package util

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"
	"time"
)

type ListingQueryFields struct {
	Origin int
	Destination int
	Date string
	Time string
}

type CreateSubmitPost struct {
	Origin int
	Destination int
	Seats int
	Fee float64
	Date string
}

type ReservationPost struct {
	ListingId int
	Seats int
}

func ValidAuthQuery(u *url.URL) (string, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", NewError(nil, "Verificación incorrecta ", 400)
	}
	if _,ok := m["t"]; !ok {
		return "", NewError(nil, "Verificación incorrecta", 400)
	}
	f := m["t"][0]
	return f, nil
}

func ValidChangePasswordQuery(u *url.URL) (string, string, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", "", NewError(nil, "Verificación incorrecta", 400)
	}
	if _,ok := m["t"]; !ok {
		return "", "", NewError(nil, "Verificación incorrecta", 400)
	}
	if _,ok := m["u"]; !ok {
		return "", "", NewError(nil, "Verificación incorrecta", 400)
	}

	return m["t"][0], m["u"][0], nil
}

func ValidChangePasswordSubmit(r *http.Request) error {
	if r.FormValue("Password") == "" || r.FormValue("Password2") == "" || r.FormValue("User") == "" || r.FormValue("Token") == "" {
		return NewError(nil, "Rellena el formulario completo por favor", 400)
	}

	if utf8.RuneCountInString(r.FormValue("Password")) < 6 {
		return NewError(nil, "La contraseña debe tener al menos 6 caracteres", 400)
	}

	if r.FormValue("Password") != r.FormValue("Password2"){
		return NewError(nil, "No coincide la contraseña", 400)
	}

	return nil
}

func ValidCreateSubmit(r *http.Request) (CreateSubmitPost, error) {
	//if the values that should be ints actually are. If not, return error.
	//Check if values are empty.
	values := CreateSubmitPost{}
	if r.FormValue("Date") == "" || r.FormValue("Time") == "" || r.FormValue("Seats") == "" || r.FormValue("Fee") == "" {
		return values, NewError(nil, "Rellena el formulario completo por favor", 400)
	}
	err := errors.New("")
	values.Origin, err = strconv.Atoi(r.FormValue("Origin"))
	if err != nil {
		return values, NewError(nil, "Origen invalido", 400)
	}

	values.Destination, err = strconv.Atoi(r.FormValue("Destination"))
	if err != nil {
		return values, NewError(nil, "Destino invalido", 400)
	}
	values.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return values, NewError(nil, "Número de cupos invalido", 400)
	}
	values.Fee, err = strconv.ParseFloat(r.FormValue("Fee"), 64)
	if err != nil {
		return values, NewError(nil, "Precio invalido", 400)
	}

	if r.FormValue("Origin") == r.FormValue("Destination") {
		return values, NewError(nil, "Por favor ingresa otro origen y destino", 400)
	}

	if values.Fee > 100 {
		return values, NewError(nil, "Precio demasiado alto. Max 100€", 400)
	}

	if values.Seats > 8 {
		return values, errors.New("Demasiados cupos. Max 8")
	}

	// Date leaving stuff
	timeVar, err := ReturnTime(r.FormValue("Date"), r.FormValue("Time"))
	if err != nil {
		return values, NewError(nil, "Fecha de salida invalida", 400)
	}

	if timeVar.Before(time.Now().Local()) {
		return values, NewError(nil, "No se puede hacer viajes en el pasado.", 400)
	}

	if timeVar.After(time.Now().Local().AddDate(0,2,0)) {
		return values, NewError(nil, "No se pueden crear eventos superiores a 2 meses", 400)
	}

	values.Date = timeVar.Format("2006-01-02 15:04:05")

	return values, nil
}

// CHANGE TO var := struct{}
func ValidListingQuery(u *url.URL) (ListingQueryFields, error) {
	// ParseQuery parses the URL-encoded query string and returns a map listing the values specified for each key.
	// ParseQuery always returns a non-nil map containing all the valid query parameters found
	urlParsed, err := url.Parse(u.String())
	if err != nil {
		return ListingQueryFields{}, NewError(nil, "URL invalido", 400)
	}

	m, err := url.ParseQuery(urlParsed.RawQuery)
	if err != nil {
		return ListingQueryFields{}, NewError(nil, "Faltan parámetros", 400)
	}
	if _,ok := m["o"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Falta origen", 400)
	}
	if _,ok := m["d"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Falta destino", 400)
	}
	if _,ok := m["t"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Falta fecha", 400)
	}
	if _,ok := m["h"]; !ok {
		return ListingQueryFields{}, NewError(nil, "Falta la hora", 400)
	}
	city1, err := strconv.Atoi(m["o"][0])
	if err != nil{
		return ListingQueryFields{}, NewError(nil, "Falta origen", 400)
	}
	city2, err := strconv.Atoi(m["d"][0])
	if err != nil{
		return ListingQueryFields{}, NewError(nil, "Falta destino", 400)
	}

	timeVar, err := ReturnTime(m["t"][0], m["h"][0])
	if err != nil {
		return ListingQueryFields{}, err
	}

	return ListingQueryFields{city1, city2, timeVar.Format("2006-01-02"), timeVar.Format("15:04")}, nil
}

func ValidRegister(r *http.Request) error {
		// POST validation
	if r.FormValue("Password") == "" || r.FormValue("Username") == "" || r.FormValue("Email") == "" {
		return NewError(nil, "Rellena el formulario completo por favor", 400)
	}

	if utf8.RuneCountInString(r.FormValue("Password")) < 6 {
		return NewError(nil, "Contraseña debe tener por lo menos 6 caracteres", 400)
	}

	if r.FormValue("Password") != r.FormValue("Password2"){
		return NewError(nil, "No coincide la contraseña", 400)
	}

	if r.FormValue("Email") != r.FormValue("Email2") {
		return NewError(nil, "Correos electrónicos diferentes", 400)
	}
	return nil
}

func ValidReservePost(r *http.Request) (ReservationPost, error) {
	reservePost := ReservationPost{}
	err := errors.New("")
	if r.FormValue("Seats") == "" || r.FormValue("Listing") == ""{
		return ReservationPost{}, NewError(nil, "Rellena el formulario completo por favor", 400)
	}
	
	reservePost.ListingId, err = strconv.Atoi(r.FormValue("Listing"))
	if err != nil {
		return ReservationPost{}, NewError(nil, "Viaje invalido", 400)
	}
	
	reservePost.Seats, err = strconv.Atoi(r.FormValue("Seats"))
	if err != nil {
		return ReservationPost{}, NewError(nil, "Número de cupos invalido", 400)
	}
	return reservePost, nil
}

func ValidReserveURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	if _,ok := m["l"]; !ok {
		return 0, NewError(nil, "Falta viaje id", 400)
	}
	listingId, err := strconv.Atoi(m["l"][0])
	if err != nil {
		return 0, NewError(nil, "Viaje invalido", 400)
	}
	return listingId, nil
}

func ValidDashQuery(u *url.URL) (int, error) {
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(nil, "URL invalido", 400)
	}
	if _,ok := m["i"]; !ok {
		return 0, NewError(nil, "URL invalido", 400)
	}
	f := m["i"][0]
	i, err := strconv.Atoi(f)
	if err != nil {
		return 0, NewError(nil, "URL invalido", 400)
	}
	return i, nil
}

func ValidMessageURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	if _,ok := m["i"]; !ok {
		return 0, NewError(nil, "Falta recipiente", 400)
	}
	userId, err := strconv.Atoi(m["i"][0])
	if err != nil {
		return 0, NewError(nil, "Falta recipiente", 400)
	}
	return userId, nil
}

func ValidMessagePost(r *http.Request) (int, string, error) {
	if r.FormValue("Recipient") == "" || r.FormValue("Message") == ""{
		return 0, "", NewError(nil, "Rellena el formulario completo por favor", 400)
	}
	
	recipient, err := strconv.Atoi(r.FormValue("Recipient"))
	if err != nil {
		return 0, "", NewError(nil, "Falta recipiente", 400)
	}
	if utf8.RuneCountInString(r.FormValue("Message")) > 500 {
		return 0, "", NewError(nil, "Mensaje demasiado largo. Max 500 caracteres", 400)
	}
	return recipient, r.FormValue("Message"), nil
}

func ValidRateURL(r *http.Request) (int, error) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return 0, NewError(err, "Error de servidor", 500)
	}
	if _,ok := m["i"]; !ok {
		return 0, NewError(nil, "URL invalido", 400)
	}
	userId, err := strconv.Atoi(m["i"][0])
	if err != nil {
		return 0, NewError(nil, "URL invalido", 400)
	}
	return userId, nil
}

func ValidRatePost(r *http.Request) (int, bool, string, bool, error) {
	if r.FormValue("User") == "" || r.FormValue("Positive") == "" {
		return 0, false, "", false, NewError(nil, "Rellena el formulario completo por favor", 400)
	}
	
	recipient, err := strconv.Atoi(r.FormValue("User"))
	if err != nil {
		return 0, false, "", false, NewError(nil, "Invalid recipient", 400)
	}
	if utf8.RuneCountInString(r.FormValue("Comment")) > 200 {
		return 0, false, "", false, NewError(nil, "Comentario demasiado largo. Max 200 caracteres", 400)
	}

	var positive, public bool
	if r.FormValue("Positive") == "true" {
		positive = true
	} else if r.FormValue("Positive") == "false" {
		positive = false
	} else {
		return 0, false, "", false, NewError(nil, "Rating invalido", 400)
	}

	if r.FormValue("Public") == "true" {
		public = true
	} else {
		positive = false
	}

	return recipient, positive, r.FormValue("Comment"), public, nil
}