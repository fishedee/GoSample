package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "time"
    "runtime"
)
func request(url string , doneChans chan int){
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(url,err)
		doneChans <- 0
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(url,resp.StatusCode);
		doneChans <- 0
		return
	}

	body,_ := ioutil.ReadAll(resp.Body)

	doneChans <- len(body)
}
func main(){
	//config
	url := "http://www.sogou.com"
	num := 500
	forkNum := 10

	startTime := time.Now().UnixNano()

	//run
	runtime.GOMAXPROCS(forkNum)
	doneChans := make( chan int )
	for i := 0 ; i < num/forkNum ; i++ {
		go request(url,doneChans)
	}

	//wait
	totalLen := 0
	for i := 0 ; i < num/forkNum ; i++ {
		totalLen += <-doneChans
	}
	
	endTime := time.Now().UnixNano()

	//result
	fmt.Println("num:",num,
		",forkNum:",forkNum,
		",totalLen:",totalLen,
		",time:",(endTime-startTime)/1e6)
}