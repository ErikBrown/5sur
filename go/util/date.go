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
func PrettyDate(timestamp string) Date {
	var splits []string = strings.SplitAfter(timestamp, "")
	var date Date
	month := splits[5] + splits[6]
	m, err := strconv.ParseInt(month, 10, 8)
	if err != nil {
		// FUCKING PANIC
	}
	date.Month = months[m-1]
	var day string
	if splits[8]=="0" {
		day = splits[9]
	} else {
		day = splits[8] + splits[9]
	}
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

func ConvertDate(d string) string {
	// We need to have a date validator somewhere
	var splits []string = strings.Split(d, "/")
	if len(splits) != 3 {
		return ""
	}
	if len(splits[0]) == 1 {
		splits[0] = "0" + splits[0]
	}
	return splits[2] + "-" + splits[1] + "-" + splits[0]
}

func ReverseConvertDate(d string) string {
	// We need to have a date validator somewhere
	var splits []string = strings.Split(d, "-")
	if len(splits) != 3 {
		return ""
	}
	if len(splits[0]) == 1 {
		splits[0] = "0" + splits[0]
	}
	return splits[2] + "/" + splits[1] + "/" + splits[0]
}