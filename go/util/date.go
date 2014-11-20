package util

import (
	"strings"
	"strconv"
	"errors"
	"math"
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

func CompareDate(d1 string, d2 string) error {
	var splitsLeaving []string = strings.Split(d1, "/")
	var splitsNow []string = strings.Split(d2, "/")
	if len(splitsLeaving) != 3 && len(splitsNow) != 3 {
		return errors.New(d2 + " & length of array: " + strconv.Itoa(len(splitsNow)))
	}
	dateLeaving := 0.0
	dateNow := 0.0
	for i := range splitsLeaving {
		leaving, err := strconv.ParseFloat(splitsLeaving[i],64)
		if err != nil {
			return err
		}
		now, err := strconv.ParseFloat(splitsNow[i],64)
		if err != nil {
			return err
		}
		dateLeaving += leaving * math.Pow(10,(math.Abs(float64(i)-2)*2))
		dateNow += now * math.Pow(10,(math.Abs(float64(i)-2)*2))
	}
	if dateLeaving < dateNow {
		return errors.New("Can't make a listing in the past joker")
	}
	return nil
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