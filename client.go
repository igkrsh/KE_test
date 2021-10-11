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
}

func CreateClient (serv * Server) * Client {
	elemLim, timeLim := serv.GetLimits()
	c := Client{
		elemLimit: elemLim,
		timeLimit: timeLim,
		serv:      serv,
	}
	return &c
}

func (cli Client) fillBatch(itemChannel chan Item, batchChannel chan Batch, timer * time.Timer) {
	var batch []Item = nil
	for {
		select {
		case item := <-itemChannel:
			batch = append(batch, item)
			if len(batch) == int(cli.elemLimit) {
				batchChannel <- batch
				batch = nil
				timer = time.NewTimer(cli.timeLimit)
			}
		case <-timer.C :
			batchChannel <- batch
			batch = nil
			timer = time.NewTimer(cli.timeLimit)
		}
	}
}

func (cli Client) sendBatch(batchChannel chan Batch, errors chan error) {
	for {
		batch := <-batchChannel
		con := context.Background()
		err := cli.serv.Process(con, batch)
		if err != nil {
			errors <- err
		}
		errors <- nil
		time.Sleep(cli.timeLimit)
	}
}

func (cli Client) addToBatch(itemChannel chan<- Item) {
	for {
		itemChannel <- Item{}
	}
}

func (cli Client) Produce() error {
	itemChannel := make(chan Item, cli.elemLimit)
	batchChannel := make(chan Batch, 5)
	errors := make(chan error)
	numOfWorkers := 10000
	timer := time.NewTimer(cli.timeLimit)
	for i := 0; i < numOfWorkers; i++ {
		go cli.addToBatch(itemChannel)
	}
	go cli.fillBatch(itemChannel, batchChannel, timer)
	go cli.sendBatch(batchChannel, errors)
	defer func() {
		close(itemChannel)
		close(batchChannel)
		close(errors)
	}()
	for {
		error := <- errors
		if error != nil {
			return error
		}
	}
}