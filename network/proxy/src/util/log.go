package util

import (
	"log"
	"os"
)

var Logger* log.Logger

func InitLogger(fileAddress string)(error){
	logfile,err := os.OpenFile(fileAddress,os.O_RDWR | os.O_APPEND|os.O_CREATE,0660);
	if err != nil{
		return err
	}

	Logger = log.New(logfile,"",log.LstdFlags)
	return nil
}
