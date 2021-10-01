package main

import (
	"time"
)

func main() {
	serv := CreateServer(10000, 5000 * time.Millisecond)
	cli := CreateClient(serv)
	cli.Schedule()
}