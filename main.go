package main

import (
	"time"
)

func main() {
	serv := CreateServer(1000000, 1000 * time.Millisecond)
	cli := CreateClient(serv)
	cli.Schedule()
}