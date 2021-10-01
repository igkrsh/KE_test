package main

import (
	"time"
)

func main() {
	serv := CreateServer(100000, 500 * time.Millisecond)
	cli := CreateClient(serv)
	cli.Send()
}