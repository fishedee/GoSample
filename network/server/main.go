package main
 
import (
	"io"
    "net/http"
    "runtime"
    "fmt"
)
 
func SayHello(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req);
    io.WriteString(w,"Hello World")
}
 
func main() {

	runtime.GOMAXPROCS( 5 )

    http.HandleFunc("/", SayHello)
    http.ListenAndServe(":3007", nil)
}