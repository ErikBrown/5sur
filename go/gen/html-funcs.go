package gen

import (
	"data/util"
	"database/sql"
	"strconv"
)

func DashListingsPage(dashListings []DashListing, listing SpecificListing, user string) (string, error) {
	title := "Dashboard - Listings"
	headerInfo := Header {
		Title: title,
		User: user,
	}

	dashListingsPage := HeaderHtml(&headerInfo)
	dashListingsPage += DashListingsHtml(dashListings, listing)
	dashListingsPage += FooterHtml()
	return dashListingsPage, nil
}
/*
func DashMessagesPage(dashMessages []DashMessages, dashMessage SpecificMessage, user string) (string, error) {
	title := "Dashboard - Messages"
	headerInfo := Header {
		Title: title,
		User: user,
	}
	dashMessagesPage := HeaderHtml(&headerInfo)
	dashMessagesPage += DashMessagesHtml(dashMessages, dashMessage)
	dashMessagesPage += FooterHtml()
	return dashMessagesPage, nil
}
*/

func HomePage(db *sql.DB, user string) (string, error) {
	title := "5Sur"
	cities, err := ReturnFilter(db)
	if err != nil { return "", err }

	headerInfo := Header {
		Title: title,
		User: user,
	}

	homePage := HeaderHtml(&headerInfo)
	homePage += FilterHtml(cities, util.ListingQueryFields{})
	homePage += FooterHtml()

	return homePage, nil
}

func CreateListingPage(db *sql.DB, user string) (string, error) {
	title := "Create Listing"

	headerInfo := Header {
		Title: title,
		User: user,
	}
	createListingPage := HeaderHtml(&headerInfo)
	cities, err := ReturnFilter(db)
	if err != nil { return "", err }

 	createListingPage += CreateListingHtml(user, cities)
	createListingPage += FooterHtml()
	return createListingPage, nil
}

func CreateReservePage(listingId int, seats int, user string, message string) string {
		// HTML generation
	headerInfo := Header {
		Title: "Reserve Page",
		User: user,
	}

	reservePage := HeaderHtml(&headerInfo)
	// Temp
	reservePage += "<br /><br /><br /><br />Placed on the reservation queue!\r\nListing ID: " + strconv.Itoa(listingId) + "\r\nSeats: " + strconv.Itoa(seats) + "User: " + user + "\r\nMessage: " + message
	reservePage += FooterHtml()
	return reservePage
}

func CreateReserveFormPage(l Listing, user string) string {
	// HTML generation
	headerInfo := Header {
		Title: "Reserve Page",
		User: user,
	}
	reservePage := HeaderHtml(&headerInfo)
	reservePage += ReserveHtml(l)
	reservePage += FooterHtml()
	return reservePage
}