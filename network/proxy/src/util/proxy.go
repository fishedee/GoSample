package util

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	//"fmt"
	"errors"
)

type RouteHandler struct{
	Host string
	Timeout time.Duration
}

func (this *RouteHandler) HandleHttp(result chan error ,writer http.ResponseWriter,request *http.Request){
	client := &http.Client{
		Timeout:time.Second * 30,
	}

	request.Host = this.Host
	request.URL.Scheme = "http"
	request.URL.Host = this.Host
	request.RequestURI = ""

	//fmt.Println(request)
    resp, err := client.Do(request)
    if err != nil{
    	result <- err
    	return
    }
    defer resp.Body.Close()

    for k, v := range resp.Header {
        for _, vv := range v {
            writer.Header().Add(k, vv)
        }
    }
    writer.WriteHeader(resp.StatusCode)
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil{
    	result <- err
    	return
    }
    writer.Write(body)
    result <- nil
}

func (this *RouteHandler) HandleTimeoutAndHttp(writer http.ResponseWriter,request *http.Request)(error){
	resultChan := make(chan error)
	var err error
	go this.HandleHttp(resultChan,writer,request)
	select {
	case result := <- resultChan:
		err = result
	case <- time.After(this.Timeout):
		err = errors.New("严重超时")
	}
	close(resultChan)
	return err
}

func (this *RouteHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request){
	beginTime := time.Now().UnixNano()
	err := this.HandleTimeoutAndHttp(writer,request)
	endTime := time.Now().UnixNano()

	var errorDesc = "";
	if err != nil{
		errorDesc = err.Error()
	}
	
	Logger.Print(
			request.RemoteAddr,
			" -- ",
			"[",
			request.Method,
			" ",
			request.RequestURI,
			" ",
			float64(endTime-beginTime)/1000000,
			"ms",
			"]",
			errorDesc,
		)
}

func SeviceProxy(port int,location []Location)(error){
	for _,singleLocation := range location{
		http.Handle(
			singleLocation.Url,
			&RouteHandler{
				Host:singleLocation.Proxy,
				Timeout:time.Duration(singleLocation.Timeout) * time.Millisecond,
			},
		)
	}
	Logger.Print("Start Proxy Server Listen On :"+strconv.Itoa(port))
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil{
		return err
	}
	return nil
}