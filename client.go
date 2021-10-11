package main

import (
	"context"
	"time"
)

type Sender interface {
	Produce() error
	addToBatch(itemChannel chan<- Item)
	sendBatch(batchChannel chan Batch, errors chan error)
	fillBatch(itemChannel chan Item, batchChannel chan Batch, timer * time.Timer)
}

type Client struct {
	elemLimit uint64
	timeLimit time.Duration
	serv * Server
	itemChannel chan Item
	batchChannel chan Batch
	errors chan error
}

func CreateClient (serv * Server) * Client {
	elemLim, timeLim := serv.GetLimits()
	c := Client{
		elemLimit:    elemLim,
		timeLimit:    timeLim,
		serv:         serv,
		itemChannel:  make(chan Item, elemLim),
		batchChannel: make(chan Batch, 5),
		errors:       make(chan error),
	}
	return &c
}

func (cli * Client) fillBatch() {
	var batch []Item = nil
	timer := time.NewTimer(cli.timeLimit)
	for {
		select {
		case item := <-cli.itemChannel:
			batch = append(batch, item)
			if len(batch) == int(cli.elemLimit) {
				cli.batchChannel <- batch
				batch = nil
				timer = time.NewTimer(cli.timeLimit)
			}
		case <-timer.C :
			cli.batchChannel <- batch
			batch = nil
			timer = time.NewTimer(cli.timeLimit)
		}
	}
}

func (cli * Client) sendBatch() {
	for {
		batch := <-cli.batchChannel
		con := context.Background()
		err := cli.serv.Process(con, batch)
		if err != nil {
			cli.errors <- err
		}
		cli.errors <- nil
		time.Sleep(cli.timeLimit)
	}
}

func (cli * Client) addToBatch() {
	for {
		cli.itemChannel <- Item{}
	}
}

func (cli * Client) Produce() error {
	numOfWorkers := 1000
	for i := 0; i < numOfWorkers; i++ {
		go cli.addToBatch()
	}
	go cli.fillBatch()
	go cli.sendBatch()
	for {
		error := <- cli.errors
		if error != nil {
			return error
		}
	}
}