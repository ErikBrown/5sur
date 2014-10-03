package util

import (
	"strings"
)

type Date struct{
	Month string
	Time string
	Day string
}

var []months = {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

//FORM yyyy-mm-dd hh:mm:ss Drop the seconds. Parse the rest.
func customDate (date string) Date{
	var splits []string = strings.SplitAfter(date, "")
	var date Date
	var month = splits[5] + splits[6]
	date.Month = months[ParseInt(month, 10, 8)-1]
	var day = splits[8] + splits[9]
	if(splits[9]=="1"){
		day+="st"
	}
	else if(splits[9]=="2"){
		day+="nd"
	} else if(splits[9]=="3"){
		day+="rd"
	}else{
		day+="th"
	}
	date.Day = day
	date.Time = splits[11] + splits[12] + ":" + splits[14] + splits[15]
	return date
}