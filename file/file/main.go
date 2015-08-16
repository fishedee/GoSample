package main

import (
	"fmt"
	"os"
)

func read(){
	file,err := os.Open("read.txt");
	if err != nil{
		fmt.Println(err);
		return;
	}

	defer file.Close();

	buf := make([]byte,1024)

	for{
		n,err := file.Read(buf);
		if err != nil{
			fmt.Println(err);
			return;
		}
		if n == 0{
			break;
		}
		fmt.Print(string(buf[:n]));
	}
}

func write(){
	file,err := os.Create("write.txt");
	if err != nil{
		fmt.Println(err);
		return;
	}

	defer file.Close();

	for i:= 0 ; i != 10 ; i++ {
		file.WriteString("Hello你好\n");
	}
}

func main(){
	
	read();

	write();
}