package main

import (
	"time"
)

func main() {
	serv := CreateServer(10000000, 500 * time.Millisecond)
	cli := CreateClient(serv)
	cli.Produce()
}