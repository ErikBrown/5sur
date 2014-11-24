package util

import (
	"strings"
	"strconv"
	"errors"
	"math"
	"time"
)

type Date struct{
	Month string
	Day string
	Time string
}

var months = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

//FORM yyyy-mm-dd hh:mm:ss Drop the seconds. Parse the rest.
func PrettyDate(timestamp string, suffix bool) Date {
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
	return date
}

// Normalizes the following to YYYY-MM-DD
// YYYY/MM/DD
// DD/MM/YYYY
// DD-MM-YYYY
// and adds 0 to single digit months/days
// This ugly code does not check for valid dates, only the format
// Return empty if in incompatible format
func ConvertDate(d string) string {
	// We need to have a date validator somewhere
	var splits []string = strings.Split(d, "/")
	if len(splits) != 3 {
		var splits []string = strings.Split(d, "-")
		if len(splits) == 3 {
			if len(splits[0]) == 4 {
				if len(splits[1]) == 1 {
					splits[1] = "0" + splits[1]
				}
				if len(splits[2]) == 1 {
					splits[2] = "0" + splits[2]
				}
				return splits[0] + "-" + splits[1] + "-" + splits[2] // Original format YYYY-MM-DD
			} else {
				if len(splits[0]) == 1 {
					splits[0] = "0" + splits[0]
				}
				if len(splits[1]) == 1 {
					splits[1] = "0" + splits[1]
				}
				return splits[2] + "-" + splits[1] + "-" + splits[0] // Original format DD-MM-YYYY
			}
		} else {
			return ""
		}
	}
	if len(splits[0]) == 1 {
		splits[0] = "0" + splits[0]
	} else if len(splits[0]) == 4 {
		if len(splits[1]) == 1 {
			splits[1] = "0" + splits[1]
		}
		if len(splits[2]) == 1 {
			splits[2] = "0" + splits[2]
		}
		return splits[0] + "-" + splits[1] + "-" + splits[2] // Original format YYYY/MM/DD
	}
	if len(splits[1]) == 1 {
		splits[1] = "0" + splits[1]
	}
	return splits[2] + "-" + splits[1] + "-" + splits[0] // Original format DD/MM/YYYY
}

func CompareDate(d1 string, d2 string) error {
	var splitsLeaving []string = strings.Split(d1, "-")
	var splitsTemp []string = strings.Split(d2, "T")
	var splitsNow []string = strings.Split(splitsTemp[0], "-")
	if len(splitsLeaving) != 3 && len(splitsNow) != 3 {
		return errors.New("Incorrect date format")
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
		return errors.New("Can't make listings in the past joker")
	}
	return nil
}

// Changes YYYY-MM-DD to DD/MM/YYYY
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

func ValidDate(d string) error {
	splitsLeaving := strings.Split(d, "-")
	if len(splitsLeaving) != 3 {
		return errors.New("Incorrect Date Format")
	}
	if len(splitsLeaving[0]) != 4 {
		return errors.New("Invalid year")
	}

	yearLeaving, err := strconv.Atoi(splitsLeaving[0])
	if err != nil {
		return errors.New("Invalid year")
	}

	monthLeaving, err := strconv.Atoi(splitsLeaving[1])
	if err != nil {
		return errors.New("Invalid month")
	}
	if monthLeaving > 12 || monthLeaving < 1 {
		return errors.New("Invalid month")
	}

	dayLeaving, err := strconv.Atoi(splitsLeaving[2])
	if err != nil {
		return errors.New("Invalid Day")
	}
	err = validDay(yearLeaving, monthLeaving, dayLeaving)
	if err != nil {
		return err
	}

	splitsTemp := strings.Split(time.Now().Local().AddDate(0,2,0).Format(time.RFC3339), "T")
	splitsNow := strings.Split(splitsTemp[0], "-")
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
	if dateLeaving > dateNow {
		return errors.New("Can't make listing later than two months in the future")
	}
	return nil
}

func validDay(y int, m int, d int) error {
	if d < 1 {
		return errors.New("Invalid day")
	}
	if m == 1 || m == 3 || m == 5 || m == 7 || m == 8 || m == 10 || m == 12 {
		if d > 30 {
			return errors.New("Invalid day")
		}
	} else {
		if m == 2 { // leap day check
			leapYear := false
			if math.Mod(float64(y), 4) != 0 {
				// Common year
			} else if math.Mod(float64(y),100) != 0 {
				leapYear = true
			} else if math.Mod(float64(y),400) !=0 {
				// Common year
			} else {
				leapYear = true
			}
			if leapYear {
				if d > 29 {
					return errors.New("Invalid day")
				}
			} else {
				if d > 28 {
					return errors.New("Invalid day")
				}
			}
		} else {
			if d > 30 {
				return errors.New("Invalid day")
			}
		}
	}
	return nil
}