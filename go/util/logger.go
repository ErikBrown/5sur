package util

import (
	"os"
	"log"
)

func ConfigureLog() {
	f, _ := os.OpenFile("error.log", os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0770)
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}