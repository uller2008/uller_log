package main

import (
	"fmt"
	uller "github.com/uller2008/uller_log"
	"math/rand"
	"strconv"
	"time"
)

func main(){
	logClientInstance := uller.LogClientInstance()
	for i:=1;i<100;i++{
		rnd := rand.Intn(10)
		go logClientInstance.Debug(strconv.Itoa(rnd),strconv.Itoa(rnd),strconv.Itoa(rnd))
		//time.Sleep(100*time.Millisecond)
	}
	fmt.Println("finished")
	time.Sleep(time.Duration(600) * time.Second)
}