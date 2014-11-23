package gen

import (
	"data/util"
	"database/sql"
	"strconv"
)

func ListingsPage(db *sql.DB, query util.ListingQueryFields, user string) (string, error) {
	title := "Listings"
	cities := ReturnFilter(db)
	listings := ReturnListings(db, query.Origin, query.Destination, query.Time)

	headerInfo := Header {
		Title: title,
		User: user,
	}

	listPage := HeaderHtml(&headerInfo)
	listPage += FilterHtml(cities, query.Origin, query.Destination, util.ReverseConvertDate(query.Time))
	listPage += ListingsHtml(listings)
	listPage += FooterHtml()

	return listPage, nil
}

func CreateListingPage(db *sql.DB, user string) (string, error) {
	title := "Create Listing"

	headerInfo := Header {
		Title: title,
		User: user,
	}
	createListingPage := HeaderHtml(&headerInfo)
	cities := ReturnFilter(db)
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

func CreateReserveFormPage(l string, user string) string {
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