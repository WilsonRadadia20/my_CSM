package logs

import (
	"log"
	"os"
)

const (
	INFO = iota
	WARNING
	ERROR
)

var (
	infoLog     *log.Logger
	wariningLog *log.Logger
	errorlog    *log.Logger
)

func init() {
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	wariningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorlog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
