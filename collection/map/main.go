package main
import (
    "fmt"
)
func main() {
    m1 := map[string]string{"a":"bb","b":"bb"}
   	fmt.Println(m1);
   
    m2 := m1;
    m2["a"] = "cc";

    fmt.Println(m1);
    fmt.Println(m2);
}