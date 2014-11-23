package gen

import (
	"data/util"
	"database/sql"
)

func ListingsPage(db *sql.DB, query util.ListingQueryFields, user string, title string) (string, error) {
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