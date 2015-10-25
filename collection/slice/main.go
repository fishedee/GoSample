package main
import (
    "fmt"
)
func main() {
    m1 := []string{"aa","bb"}
    fmt.Println(m1);

    m2 := m1;
    m2[0] = "cc";

    fmt.Println(m1);
    fmt.Println(m2);
}