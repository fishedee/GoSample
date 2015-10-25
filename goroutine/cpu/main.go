package main
import (
    "fmt"
    "runtime"
    "time"
)
func test(i int,max int)int{
    return i%max;
}
func say(c chan bool,dd string) {
	max := 1000000;
	for i := 0 ; i <= max ; i++ {
		i = test(i,max);
		//fmt.Println(dd);
	}
	//c <- true;
}
func temp(c chan bool,quit chan bool){
    runtime.LockOSThread()
    fmt.Println(time.Now());
    timeout := time.After(time.Second);
        select {
            case <-timeout:
                fmt.Println("timeout");
                break;
            default:
                fmt.Println("tick!!!");
                time.Sleep(time.Second);
        }
    
    fmt.Println(time.Now());
    fmt.Println("exit!");
    //quit <- true
}
func main() {
    runtime.GOMAXPROCS(1)
	c := make(chan bool);
	quit := make(chan bool)
    go say(c,"A");
    go say(c,"B");
	temp(c,quit);
    //<- quit
}