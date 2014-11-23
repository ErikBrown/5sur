package gen

import (
	"data/util"
	"database/sql"
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