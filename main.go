package main

import (
	"time"
)

func main() {
	serv := CreateServer(100000, 50 * time.Millisecond)
	cli := CreateClient(serv)
	cli.Schedule()
}