package main
 
import (
	"io"
    "net/http"
    "runtime"
)
 
func SayHello(w http.ResponseWriter, req *http.Request) {
    io.WriteString(w,"Hello World")
}
 
func main() {

	runtime.GOMAXPROCS( 5 )

    http.HandleFunc("/", SayHello)
    http.ListenAndServe(":8888", nil)
}