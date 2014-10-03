package main

import (
	"net/http"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"data/results"
	"data/util"
)

func createHTML (myListing results.Listing) string{
	var date util.Date = util.CustomDate(myListing.DateLeaving)
	output := "<ul class=\"list_item\"><li class=\"listing_user\"><img src=\"http://192.241.219.35/" + myListing.Picture + "\" alt=\"User Picture\"><span class=\"positive\">+100</span></li><li class=\"date_leaving\"><div><span class=\"month\">" + date.Month + "</span><span class=\"month\">" + date.Day + "</span><span class=\"month\">" + date.Time + "</span></div></li><li class=\"city\"><span>" + myListing.Origin + "</span><span class=\"to\">&#10132;</span><span>" + myListing.Destination + "</span></li><li class=\"seats\"><span>2</span></li><li class=\"fee\"><span>$" + fmt.Sprintf("%.2f", myListing.Fee) + "</span></li></ul>"
	return output
}

func generateHtml(w http.ResponseWriter, r *http.Request) {
	results := results.ReturnListings() // Make struct to store everything
	myString := "<!doctype html><html><head><title>Title</title><link href='http://fonts.googleapis.com/css?family=Montserrat:400,700|Open+Sans:400,400italic,600,300,700,800|Bitter:400,400italic,700' rel='stylesheet' type='text/css'><link rel=\"stylesheet\" type=\"text/css\" href=\"http://192.241.219.35/style.css\" /></head><body>"
	for i := range results{
		myString += createHTML(results[i])
	}
	myString += "</body></html>"
	fmt.Fprint(w, myString)
}

func main() {
	http.HandleFunc("/go/", generateHtml)
	http.ListenAndServe(":8080", nil)
}