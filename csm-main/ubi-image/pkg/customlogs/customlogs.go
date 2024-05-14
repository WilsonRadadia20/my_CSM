package customlogs

import (
	"log"
	"os"
)

var (
	InfoLog     *log.Logger
	WariningLog *log.Logger
	Errorlog    *log.Logger
)

func init() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	WariningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime)
	Errorlog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
}
