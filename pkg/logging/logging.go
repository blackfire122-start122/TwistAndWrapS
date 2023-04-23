package logging

import (
	"io"
	"log"
	"os"
)

var (
	ErrorLogger *log.Logger
)

func init() {
	errFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	ErrorLogger = log.New(io.MultiWriter(errFile, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
