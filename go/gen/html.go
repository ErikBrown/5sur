package gen

import (
	"data/util"
	"html/template"
)

type LoginHTML struct {
	Title string
	Script template.HTML
	Captcha template.HTML
}

type HeaderHTML struct {
	Title string
	Username string
	Alerts int
	AlertText []template.HTML
	UserImage string
}

type DashMessagesHTML struct {
	SidebarMessages []DashMessages
	MessageThread MessageThread
}

type DashListingsHTML struct {
	SidebarListings []DashListing
	Listing SpecificListing
}

type DashReservationsHTML struct {
	SidebarReservations []DashReservation
	Reservation Reservation
}

type ListingsHTML struct {
	Filter []City
	Listings []Listing
	Query util.ListingQueryFields
}

type ReserveHTML struct {
	ListingId int
	Driver string
	Seats []int
}

type Header struct {
	Title string
	User string
	Messages int
}

type City struct {
	Id int
	Name string
}

type Listing struct {
	Id int
	Driver int
	Picture string
	Timestamp string
	Date string
	Time string
	Origin string
	Destination string
	Seats int
	Fee float32
}