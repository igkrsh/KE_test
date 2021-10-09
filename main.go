package main

import (
	"time"
)

func main() {
	serv := CreateServer(1000000, 5000 * time.Millisecond)
	cli := CreateClient(serv)
	for i := 0; i < 10; i ++ {
		cli.Schedule()
	}
}