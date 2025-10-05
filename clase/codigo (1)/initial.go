package main

import (
	"fmt"
	"time"
)

func extra(){
	for i := 0; i < 5; i++ {
		fmt.Println("[°] Aux")	
		time.Sleep(2 * time.Second)		
	}	
}

func main(){
	var i int
	var j string = "a"
	var w float32 = 1.5
	var z bool = false
	go extra()
	fmt.Println("[°] WORK")
	fmt.Print("[°] Ingresa valor: ")
	fmt.Scan(&i)
	fmt.Println(i,j,w,z)
}