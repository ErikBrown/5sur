package util

import (
	"strings"
	"strconv"
	"time"
	"unicode/utf8"
)

type Date struct{
	Month string
	Day string
	Time string
}

var months = [12]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

// Determines time layout for the following:
// YYYY-MM-DD, MM-DD-YYYY, MM-DD-YY (the punctuation is variable)
// Month and Day can be single digits
// The punctuation (-,/,_,etc) is determined by the second parameter
func parseTimeLayout(s []string, p string) string {
	if len(s) == 3 {
		if utf8.RuneCountInString(s[0]) == 4 {
			return "2006" + p + "1" + p + "_2"
		} else {
			if utf8.RuneCountInString(s[2]) == 4 {
				return "_2" + p + "1" + p + "2006"
			} else if utf8.RuneCountInString(s[2])== 2 {
				return "_2" + p + "1" + p + "06"
			}
		}
	}
	return ""
}

// Returns time layout for time.Parse
func returnTimeLayout(t string) string {
	splits := strings.Split(t, "/")
	layout := parseTimeLayout(splits, "/")
	if layout == "" {
		splits = strings.Split(t, "-")
		layout = parseTimeLayout(splits, "-")
	}
	if layout == "" {
		return "1-_2-2006"
	} else {
		return layout
	}
}

// Takes year/month/day (in a variety of formats) and HH:MM parameters
// Checks if the time is valid and returns error if it is not
// Parses in the location of Santiago
func ReturnTime(d string, t string) (time.Time, error) {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return time.Time{}, NewError(err, "Internal server error", 500)
	}

	layout := returnTimeLayout(d)
	splits := strings.Split(t, ":")
	if len(splits) == 1 {
		if utf8.RuneCountInString(t) == 2 {
			layout += " 15"
		} else if utf8.RuneCountInString(t) == 1 {
			t = "0" + t
			layout += " 15"
		} else if utf8.RuneCountInString(t) == 4{
			layout += " 1504"
		}
	} else if len(splits) == 2 {
		layout += " 15:04"
	}

	timeVar, err := time.ParseInLocation(layout, d + " " + t, loc)
	if err != nil {
		return time.Time{}, NewError(nil, "Invalid time format", 400)
	}

	return timeVar, nil
}

// Normalizes time format to one of two layouts (machine or human readable)
// Checks if the time is valid and returns error if it is not
// Parses in the location of Santiago
func ReturnTimeString(humanReadable bool, d string, t string) (string, string, error) {
	const (
		layoutHuman = "2/1/2006"
		layoutMachine = "2006-01-02"
	)
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return "", "", NewError(err, "Internal server error", 500)
	}

	layout := returnTimeLayout(d)
	splits := strings.Split(t, ":")
	if len(splits) == 1 {
		if utf8.RuneCountInString(t) == 2 {
			layout += " 15"
		} else if utf8.RuneCountInString(t) == 1 {
			t = "0" + t
			layout += " 1"
		} else if utf8.RuneCountInString(t) == 4{
			layout += " 1504"
		}
	} else if len(splits) == 2 {
		layout += " 15:04"
	}

	timeVar, err := time.ParseInLocation(layout, d + " " + t, loc)
	if err != nil {
		return "", "", NewError(nil, "Invalid time format", 400)
	}
	if humanReadable {
		return timeVar.Format(layoutHuman), timeVar.Format("15:04"), nil
	} else {
		return timeVar.Format(layoutMachine), timeVar.Format("15:04"), nil
	}
}

func ReturnCurrentTimeString(rounded bool) (string, string) {
	if rounded {
		return time.Now().Local().Format("2006-01-02"), time.Now().Local().Format("15") + ":00"
	} else {
		return time.Now().Local().Format("2006-01-02"), time.Now().Local().Format("15:04")
	}
}

//FORM yyyy-mm-dd hh:mm:ss Drop the seconds. Parse the rest.
func PrettyDate(timestamp string, suffix bool) (Date, error) {
	var splits []string = strings.Split(timestamp, "")
	var date Date
	month := splits[5] + splits[6]
	m, err := strconv.ParseInt(month, 10, 8)
	if err != nil {
		return date, NewError(err, "Internal server error", 500)
	}
	date.Month = months[m-1]
	var day string
	if splits[8]=="0" {
		day = splits[9]
	} else {
		day = splits[8] + splits[9]
	}
	if suffix {
		if splits[9]=="1" {
			day+="st"
		} else if splits[9]=="2" {
			day+="nd"
		} else if splits[9]=="3" {
			day+="rd"
		} else{
			day+="th"
		}
	}
	date.Day = day
	date.Time = splits[11] + splits[12] + ":" + splits[14] + splits[15]
	return date, nil
}

func TimeStringInPast(t string) (bool, error) {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		return false, NewError(err, "Internal server error", 500)
	}
	timeVar, err := time.ParseInLocation("2006-01-02 15:04:05", t, loc)
	if err != nil {
		return false, NewError(nil, "Invalid time format", 400)
	}

	return timeVar.Before(time.Now()), nil
}