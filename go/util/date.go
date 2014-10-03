package util

import (
	"strings"
	"strconv"
)

type Date struct{
	Month string
	Day string
	Time string
}

var months = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

//FORM yyyy-mm-dd hh:mm:ss Drop the seconds. Parse the rest.
func CustomDate (timestamp string) Date{
	var splits []string = strings.SplitAfter(timestamp, "")
	var date Date
	month := splits[5] + splits[6]
	m, err := strconv.ParseInt(month, 10, 8)
	if err != nil {
		// FUCKING PANIC
	}
	date.Month = months[m-1]
	var day = splits[8] + splits[9]
	if splits[9]=="1" {
		day+="st"
	} else if splits[9]=="2" {
		day+="nd"
	} else if splits[9]=="3" {
		day+="rd"
	} else{
		day+="th"
	}
	date.Day = day
	date.Time = splits[11] + splits[12] + ":" + splits[14] + splits[15]
	return date
}