package util

import (
	"os"
	"log"
	"runtime"
	"strconv"
)

// If we do not want to log the error, set LogError to nil
type MyError struct {
	LogError error
	LogErrorPos string
	Message string
	StatusCode int
}

func (e MyError) Error() string {
	if e.LogError != nil { return e.LogError.Error() }
	return ""
}

func errorPosition() string {
	pc, _, line, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name() + " " + strconv.Itoa(line)
}

func NewError(err error, message string, code int) error {
	return MyError{err, errorPosition(), message, code}
}

func ConfigureLog() {
	f, _ := os.OpenFile("error.log", os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0770)
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime)
}

func PrintLog(err MyError) {
	log.Println(err.LogErrorPos + ": " + err.Error() + "\r\n \r\n")
}