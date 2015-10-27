package util

import (
	"io/ioutil"
	"os"
	"encoding/json"
)

type Location struct{
	Url string `json:"url"`
	Proxy string `json:"proxy"`
	Timeout int `json:"timeout"`
}

type Config struct{
	Listen int `json:"listen"`
	LogFile string `json:"log_file"`
	Location []Location `json:"location"`
}

func GetConfigFromFile(fileName string)(*Config,error){
	file, err := os.Open(fileName)
    if err != nil {
    	return nil,err
    }
    defer file.Close()

    configFile,err := ioutil.ReadAll(file)
    if err != nil{
    	return nil,err
    }

    var result *Config;
	err = json.Unmarshal(configFile,&result);
	if err != nil{
		return nil,err
	}

	return result,err
}