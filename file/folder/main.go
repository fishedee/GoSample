package main
import (
    "path/filepath"
    "os"
    "fmt"
)

func getFilelist(path string) {
    err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
            if ( f == nil ) {return err}
            println(path)
            return nil
    })
    if err != nil {
            fmt.Println(err)
    }
}

func main(){
    getFilelist("../../")
}