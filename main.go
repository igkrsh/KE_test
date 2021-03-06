package main

import (
	"fmt"
	"time"
)

func main() {
	serv := CreateServer(10000000, 500 * time.Millisecond)
	cli := CreateClient(serv)
	err := cli.Produce()
	if err != nil {
		fmt.Println(err)
	}
}